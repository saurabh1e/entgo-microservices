package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/saurabh/entgo-microservices/auth/config"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/permission"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/role"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/rolepermission"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/user"
	"github.com/saurabh/entgo-microservices/auth/utils/database"

	"github.com/saurabh/entgo-microservices/pkg/authz"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"

	_ "github.com/saurabh/entgo-microservices/auth/internal/ent/runtime"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logConfig := logger.LogConfig{
		Level:      cfg.Logging.Level,
		LogDir:     cfg.Logging.LogDir,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
	if err := logger.InitLogger(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("🌱 Starting database seeding...")

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.WithError(err).Error("Failed to close database connection")
		}
	}()

	client := db.Client
	ctx := context.Background()

	// Bypass authorization checks for seeding - no user context available during seeding
	ctx = authz.SetBypass(ctx, true)

	// Start seeding
	if err := seedDatabase(ctx, client); err != nil {
		logger.WithError(err).Fatal("Failed to seed database")
	}

	logger.Info("✅ Database seeding completed successfully!")
}

func seedDatabase(ctx context.Context, client *ent.Client) error {
	// Use transaction for atomicity
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				logger.WithError(err).Error("Failed to rollback transaction")
			}
			panic(v)
		}
	}()

	// 1. Create Admin Role
	logger.Info("Creating admin role...")
	adminRole, err := createAdminRole(ctx, tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.WithError(rbErr).Error("Failed to rollback transaction")
		}
		return fmt.Errorf("failed to create admin role: %w", err)
	}
	logger.WithField("role_id", adminRole.ID).Info("✓ Admin role created")

	// 2. Create Permissions
	logger.Info("Creating permissions...")
	permissions, err := createPermissions(ctx, tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.WithError(rbErr).Error("Failed to rollback transaction")
		}
		return fmt.Errorf("failed to create permissions: %w", err)
	}
	logger.WithField("count", len(permissions)).Info("✓ Permissions created")

	// 3. Assign Permissions to Admin Role
	logger.Info("Assigning permissions to admin role...")
	if err := assignPermissionsToRole(ctx, tx, adminRole, permissions); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.WithError(rbErr).Error("Failed to rollback transaction")
		}
		return fmt.Errorf("failed to assign permissions to role: %w", err)
	}
	logger.Info("✓ Permissions assigned to admin role")

	// 4. Create Admin User
	logger.Info("Creating admin user...")
	adminUser, err := createAdminUser(ctx, tx, adminRole)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.WithError(rbErr).Error("Failed to rollback transaction")
		}
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	logger.WithFields(map[string]interface{}{
		"user_id":  adminUser.ID,
		"username": adminUser.Username,
		"email":    adminUser.Email,
	}).Info("✓ Admin user created")

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Print summary
	printSummary(adminUser, adminRole, permissions)

	return nil
}

func createAdminRole(ctx context.Context, tx *ent.Tx) (*ent.Role, error) {
	// Check if role already exists
	existing, err := tx.Role.Query().
		Where(role.NameEQ("admin")).
		Only(ctx)
	if err == nil {
		logger.Info("Admin role already exists, using existing role")
		return existing, nil
	}

	// Create new admin role
	roleEntity, err := tx.Role.Create().
		SetName("admin").
		SetDisplayName("Administrator").
		SetDescription("Full system access with all permissions").
		SetIsActive(true).
		SetPriority(100).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return roleEntity, nil
}

func createPermissions(ctx context.Context, tx *ent.Tx) ([]*ent.Permission, error) {
	permissionsData := []struct {
		Name        string
		DisplayName string
		Description string
		Resource    string
	}{
		{
			Name:        "users.manage",
			DisplayName: "Manage Users",
			Description: "Full access to create, read, update, and delete users",
			Resource:    "users",
		},
		{
			Name:        "roles.manage",
			DisplayName: "Manage Roles",
			Description: "Full access to create, read, update, and delete roles",
			Resource:    "roles",
		},
		{
			Name:        "permissions.manage",
			DisplayName: "Manage Permissions",
			Description: "Full access to create, read, update, and delete permissions",
			Resource:    "permissions",
		},
		{
			Name:        "system.configure",
			DisplayName: "Configure System",
			Description: "Access to system configuration and settings",
			Resource:    "system",
		},
		{
			Name:        "audit.view",
			DisplayName: "View Audit Logs",
			Description: "Access to view system audit logs and reports",
			Resource:    "audit",
		},
	}

	permissions := make([]*ent.Permission, 0, len(permissionsData))

	for i, data := range permissionsData {
		// Check if permission already exists
		existing, err := tx.Permission.Query().
			Where(permission.NameEQ(data.Name)).
			Only(ctx)
		if err == nil {
			logger.WithField("name", data.Name).Info("Permission already exists, using existing permission")
			permissions = append(permissions, existing)
			continue
		}

		// Create new permission
		perm, err := tx.Permission.Create().
			SetName(data.Name).
			SetDisplayName(data.DisplayName).
			SetDescription(data.Description).
			SetResource(data.Resource).
			SetIsActive(true).
			SetCreatedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to create permission %d: %w", i+1, err)
		}

		permissions = append(permissions, perm)
		logger.WithFields(map[string]interface{}{
			"number": i + 1,
			"name":   data.Name,
		}).Info("  ✓ Permission created")
	}

	return permissions, nil
}

func assignPermissionsToRole(ctx context.Context, tx *ent.Tx, roleEntity *ent.Role, permissions []*ent.Permission) error {
	for _, perm := range permissions {
		// Check if role-permission relationship already exists
		count, err := tx.RolePermission.Query().
			Where(
				rolepermission.HasRoleWith(role.IDEQ(roleEntity.ID)),
				rolepermission.HasPermissionWith(permission.IDEQ(perm.ID)),
			).
			Count(ctx)

		if err != nil {
			return fmt.Errorf("failed to check existing role-permission: %w", err)
		}

		if count > 0 {
			logger.WithFields(map[string]interface{}{
				"role":       roleEntity.Name,
				"permission": perm.Name,
			}).Info("  ✓ Role-permission relationship already exists")
			continue
		}

		// Create role-permission relationship with full CRUD access
		_, err = tx.RolePermission.Create().
			SetRole(roleEntity).
			SetPermission(perm).
			SetCanRead(true).
			SetCanCreate(true).
			SetCanUpdate(true).
			SetCanDelete(true).
			SetCreatedAt(time.Now()).
			SetUpdatedAt(time.Now()).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to assign permission %s to role: %w", perm.Name, err)
		}

		logger.WithFields(map[string]interface{}{
			"permission": perm.Name,
			"role":       roleEntity.Name,
		}).Info("  ✓ Permission assigned to role")
	}

	return nil
}

func createAdminUser(ctx context.Context, tx *ent.Tx, roleEntity *ent.Role) (*ent.User, error) {
	// Check if admin user already exists
	existing, err := tx.User.Query().
		Where(
			user.Or(
				user.UsernameEQ("admin"),
				user.EmailEQ("admin@example.com"),
			),
		).
		Only(ctx)
	if err == nil {
		logger.Info("Admin user already exists, using existing user")
		return existing, nil
	}

	// Hash the password
	defaultPassword := "admin123" // Change this in production!
	hashedPassword, err := jwt.HashPassword(defaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	userEntity, err := tx.User.Create().
		SetEmail("admin@example.com").
		SetUsername("admin").
		SetPasswordHash(hashedPassword).
		SetName("System Administrator").
		SetUserType("admin").
		SetIsActive(true).
		SetEmailVerified(true).
		SetEmailVerifiedAt(time.Now()).
		SetRoleRef(roleEntity).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return userEntity, nil
}

func printSummary(user *ent.User, role *ent.Role, permissions []*ent.Permission) {
	fmt.Println("\n" + "═════════════════════════════════════════════════════════════")
	fmt.Println("                    SEEDING SUMMARY")
	fmt.Println("═════════════════════════════════════════════════════════════")
	fmt.Printf("\n✓ Admin User Created:\n")
	fmt.Printf("  • ID:       %d\n", user.ID)
	fmt.Printf("  • Username: %s\n", user.Username)
	fmt.Printf("  • Email:    %s\n", user.Email)
	fmt.Printf("  • Password: admin123 (CHANGE IN PRODUCTION!)\n")
	fmt.Printf("  • Name:     %s\n", user.Name)
	fmt.Printf("  • Type:     %s\n", user.UserType)
	fmt.Printf("  • Active:   %t\n", user.IsActive)

	fmt.Printf("\n✓ Admin Role Created:\n")
	fmt.Printf("  • ID:       %d\n", role.ID)
	fmt.Printf("  • Name:     %s\n", role.Name)
	fmt.Printf("  • Display:  %s\n", role.DisplayName)
	fmt.Printf("  • Priority: %d\n", role.Priority)

	fmt.Printf("\n✓ Permissions Created (%d):\n", len(permissions))
	for i, perm := range permissions {
		fmt.Printf("  %d. %s\n", i+1, perm.Name)
		fmt.Printf("     - Display: %s\n", perm.DisplayName)
		fmt.Printf("     - Resource: %s\n", perm.Resource)
		fmt.Printf("     - Access: Read, Create, Update, Delete\n")
	}

	fmt.Println("\n═════════════════════════════════════════════════════════════")
	fmt.Println("             🎉 All data seeded successfully! 🎉")
	fmt.Println("═════════════════════════════════════════════════════════════\n")
}
