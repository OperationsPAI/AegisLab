package consts

import (
	"fmt"
	"slices"
	"strings"
)

// ActionName is the type for permission action names, used for permission checks
type ActionName string

// System permission action constants
const (
	// Basic CRUD Operations
	ActionCreate ActionName = "create" // Create new resources
	ActionRead   ActionName = "read"   // Read/View resources
	ActionUpdate ActionName = "update" // Update existing resources
	ActionDelete ActionName = "delete" // Delete resources

	// Execution and State Management
	ActionExecute  ActionName = "execute"  // Execute tasks, injections, builds, etc.
	ActionStop     ActionName = "stop"     // Stop running tasks or executions
	ActionRestart  ActionName = "restart"  // Restart services or tasks
	ActionActivate ActionName = "activate" // Activate/Enable resources
	ActionSuspend  ActionName = "suspend"  // Suspend/Disable resources

	// File and Data Operations
	ActionUpload   ActionName = "upload"   // Upload files, charts, datasets
	ActionDownload ActionName = "download" // Download files, results, datapacks
	ActionImport   ActionName = "import"   // Import data, configurations
	ActionExport   ActionName = "export"   // Export data, reports, backups

	// Permission and Membership Management
	ActionAssign ActionName = "assign" // Assign users to resources, roles to users
	ActionGrant  ActionName = "grant"  // Grant permissions or access rights
	ActionRevoke ActionName = "revoke" // Revoke permissions or access rights

	// Configuration and Administration
	ActionConfigure ActionName = "configure" // Configure system settings, resources
	ActionManage    ActionName = "manage"    // Full administrative control

	// Collaboration and Sharing
	ActionShare ActionName = "share" // Share resources with others
	ActionClone ActionName = "clone" // Clone/Duplicate resources

	// Monitoring and Analysis
	ActionMonitor ActionName = "monitor" // Monitor system metrics, traces
	ActionAnalyze ActionName = "analyze" // Analyze results, evaluate performance
	ActionAudit   ActionName = "audit"   // View audit logs and history
)

func (a ActionName) String() string {
	return string(a)
}

// ResourceName is the type for resource names, used for permission checks
type ResourceName string

// System resource name constants
const (
	ResourceSystem           ResourceName = "system"            // system resource
	ResourceAudit            ResourceName = "audit"             // audit resource
	ResourceConfiguration    ResourceName = "configuration"     // configuration resource
	ResourceContainer        ResourceName = "container"         // container resource
	ResourceContainerVersion ResourceName = "container_version" // container version resource
	ResourceDataset          ResourceName = "dataset"           // dataset resource
	ResourceDatasetVersion   ResourceName = "dataset_version"   // dataset version resource
	ResourceProject          ResourceName = "project"           // project resource
	ResourceTeam             ResourceName = "team"              // team resource
	ResourceLabel            ResourceName = "label"             // label resource
	ResourceUser             ResourceName = "user"              // user resource
	ResourceRole             ResourceName = "role"              // role resource
	ResourcePermission       ResourceName = "permission"        // permission resource
	ResourceTask             ResourceName = "task"              // task resource
	ResourceTrace            ResourceName = "trace"             // trace resource
	ResourceInjection        ResourceName = "injection"         // fault injection resource
	ResourceExecution        ResourceName = "execution"         // execution resource
)

func (r ResourceName) String() string {
	return string(r)
}

var ProjectScopedResources = []ResourceName{
	ResourceInjection,
	ResourceExecution,
	ResourceTask,
	ResourceTrace,
}

var TeamScopedResources = []ResourceName{
	ResourceContainer,
	ResourceContainerVersion,
	ResourceDataset,
	ResourceDatasetVersion,
	ResourceProject,
	ResourceLabel,
}

// IsProjectScoped checks if a resource can inherit project-level permissions
func IsProjectScoped(resource ResourceName) bool {
	return slices.Contains(ProjectScopedResources, resource)
}

// IsTeamScoped checks if a resource can inherit team-level permissions
func IsTeamScoped(resource ResourceName) bool {
	return slices.Contains(TeamScopedResources, resource)
}

type ResourceType int

const (
	ResourceTypeSystem ResourceType = iota
	ResourceTypeTable
)

type ResourceCategory int

const (
	// ResourceCategoryChaos represents core chaos engineering resources
	// (e.g., injection, execution, trace, task).
	ResourceCategoryChaos ResourceCategory = iota

	// ResourceCategoryAsset represents experimental assets
	// (e.g., container, dataset, label).
	ResourceCategoryAsset

	// ResourceCategoryPlatform represents organizational and administrative structures
	// (e.g., project, team).
	ResourceCategoryPlatform

	// ResourceCategorySystem represents system configurations and security
	// (e.g., system, audit, configuration, user, role, permission).
	ResourceCategorySystem
)

type ResourceScope string

const (
	ScopeOwn     ResourceScope = "own"     // Own resources only
	ScopeProject ResourceScope = "project" // Project resources
	ScopeTeam    ResourceScope = "team"    // Team resources
	ScopeAll     ResourceScope = "all"     // All resources
)

func (r ResourceScope) String() string {
	return string(r)
}

// RoleName is the type for role constants
type RoleName string

// Role constants for system roles
const (
	RoleSuperAdmin           RoleName = "super_admin"            // Super Admin
	RoleAdmin                RoleName = "admin"                  // Admin
	RoleUser                 RoleName = "user"                   // Regular User
	RoleContainerAdmin       RoleName = "container_admin"        // Container Admin
	RoleContainerDeveloper   RoleName = "container_developer"    // Container Developer
	RoleContainerViewer      RoleName = "container_viewer"       // Container Viewer
	RoleDatasetAdmin         RoleName = "dataset_admin"          // Dataset Admin
	RoleDatasetDeveloper     RoleName = "dataset_developer"      // Dataset Developer
	RoleDatasetViewer        RoleName = "dataset_viewer"         // Dataset Viewer
	RoleProjectAdmin         RoleName = "project_admin"          // Project Admin
	RoleProjectAlgoDeveloper RoleName = "project_algo_developer" // Project Algorithm Developer
	RoleProjectDataDeveloper RoleName = "project_data_developer" // Project Data Developer
	RoleProjectViewer        RoleName = "project_viewer"         // Project Viewer
	RoleTeamAdmin            RoleName = "team_admin"             // Team Admin
	RoleTeamMember           RoleName = "team_member"            // Team Member
	RoleTeamViewer           RoleName = "team_viewer"            // Team Viewer
)

func (r RoleName) String() string {
	return string(r)
}

type PermissionRule struct {
	Resource ResourceName
	Action   ActionName
	Scope    ResourceScope
}

// ParsePermissionRule parses a permission rule string (e.g., "container:read:own")
func ParsePermissionRule(rule string) (*PermissionRule, error) {
	parts := strings.Split(rule, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid permission rule format: %s (expected resource:action:scope)", rule)
	}
	return &PermissionRule{
		Resource: ResourceName(parts[0]),
		Action:   ActionName(parts[1]),
		Scope:    ResourceScope(parts[2]),
	}, nil
}

func (pr PermissionRule) String() string {
	return fmt.Sprintf("%v:%v:%v", pr.Resource, pr.Action, pr.Scope)
}

// ============================================================================
// Predefined Permission Rules - Single Source of Truth
// ============================================================================

var (
	// System permissions
	PermSystemRead      = PermissionRule{Resource: ResourceSystem, Action: ActionRead, Scope: ScopeAll}
	PermSystemConfigure = PermissionRule{Resource: ResourceSystem, Action: ActionConfigure, Scope: ScopeAll}
	PermSystemManage    = PermissionRule{Resource: ResourceSystem, Action: ActionManage, Scope: ScopeAll}

	// Audit permissions
	PermAuditRead  = PermissionRule{Resource: ResourceAudit, Action: ActionRead, Scope: ScopeAll}
	PermAuditAudit = PermissionRule{Resource: ResourceAudit, Action: ActionAudit, Scope: ScopeAll}

	// Configuration permissions
	PermConfigurationRead      = PermissionRule{Resource: ResourceConfiguration, Action: ActionRead, Scope: ScopeAll}
	PermConfigurationUpdate    = PermissionRule{Resource: ResourceConfiguration, Action: ActionUpdate, Scope: ScopeAll}
	PermConfigurationConfigure = PermissionRule{Resource: ResourceConfiguration, Action: ActionConfigure, Scope: ScopeAll}

	// User permissions
	PermUserReadAll   = PermissionRule{Resource: ResourceUser, Action: ActionRead, Scope: ScopeAll}
	PermUserCreateAll = PermissionRule{Resource: ResourceUser, Action: ActionCreate, Scope: ScopeAll}
	PermUserUpdateAll = PermissionRule{Resource: ResourceUser, Action: ActionUpdate, Scope: ScopeAll}
	PermUserDeleteAll = PermissionRule{Resource: ResourceUser, Action: ActionDelete, Scope: ScopeAll}
	PermUserAssignAll = PermissionRule{Resource: ResourceUser, Action: ActionAssign, Scope: ScopeAll}

	// Role permissions
	PermRoleReadAll   = PermissionRule{Resource: ResourceRole, Action: ActionRead, Scope: ScopeAll}
	PermRoleCreateAll = PermissionRule{Resource: ResourceRole, Action: ActionCreate, Scope: ScopeAll}
	PermRoleUpdateAll = PermissionRule{Resource: ResourceRole, Action: ActionUpdate, Scope: ScopeAll}
	PermRoleDeleteAll = PermissionRule{Resource: ResourceRole, Action: ActionDelete, Scope: ScopeAll}
	PermRoleGrantAll  = PermissionRule{Resource: ResourceRole, Action: ActionGrant, Scope: ScopeAll}
	PermRoleRevokeAll = PermissionRule{Resource: ResourceRole, Action: ActionRevoke, Scope: ScopeAll}

	// Permission management permissions
	PermPermissionReadAll   = PermissionRule{Resource: ResourcePermission, Action: ActionRead, Scope: ScopeAll}
	PermPermissionCreateAll = PermissionRule{Resource: ResourcePermission, Action: ActionCreate, Scope: ScopeAll}
	PermPermissionUpdateAll = PermissionRule{Resource: ResourcePermission, Action: ActionUpdate, Scope: ScopeAll}
	PermPermissionDeleteAll = PermissionRule{Resource: ResourcePermission, Action: ActionDelete, Scope: ScopeAll}
	PermPermissionManageAll = PermissionRule{Resource: ResourcePermission, Action: ActionManage, Scope: ScopeAll}

	// Team permissions
	PermTeamReadAll    = PermissionRule{Resource: ResourceTeam, Action: ActionRead, Scope: ScopeAll}
	PermTeamReadTeam   = PermissionRule{Resource: ResourceTeam, Action: ActionRead, Scope: ScopeTeam}
	PermTeamCreateAll  = PermissionRule{Resource: ResourceTeam, Action: ActionCreate, Scope: ScopeAll}
	PermTeamUpdateAll  = PermissionRule{Resource: ResourceTeam, Action: ActionUpdate, Scope: ScopeAll}
	PermTeamUpdateTeam = PermissionRule{Resource: ResourceTeam, Action: ActionUpdate, Scope: ScopeTeam}
	PermTeamDeleteAll  = PermissionRule{Resource: ResourceTeam, Action: ActionDelete, Scope: ScopeAll}
	PermTeamManageAll  = PermissionRule{Resource: ResourceTeam, Action: ActionManage, Scope: ScopeAll}

	// Project permissions
	PermProjectReadAll    = PermissionRule{Resource: ResourceProject, Action: ActionRead, Scope: ScopeAll}
	PermProjectReadTeam   = PermissionRule{Resource: ResourceProject, Action: ActionRead, Scope: ScopeTeam}
	PermProjectReadOwn    = PermissionRule{Resource: ResourceProject, Action: ActionRead, Scope: ScopeOwn}
	PermProjectCreateTeam = PermissionRule{Resource: ResourceProject, Action: ActionCreate, Scope: ScopeTeam}
	PermProjectCreateOwn  = PermissionRule{Resource: ResourceProject, Action: ActionCreate, Scope: ScopeOwn}
	PermProjectUpdateAll  = PermissionRule{Resource: ResourceProject, Action: ActionUpdate, Scope: ScopeAll}
	PermProjectUpdateOwn  = PermissionRule{Resource: ResourceProject, Action: ActionUpdate, Scope: ScopeOwn}
	PermProjectDeleteAll  = PermissionRule{Resource: ResourceProject, Action: ActionDelete, Scope: ScopeAll}
	PermProjectDeleteOwn  = PermissionRule{Resource: ResourceProject, Action: ActionDelete, Scope: ScopeOwn}
	PermProjectManageAll  = PermissionRule{Resource: ResourceProject, Action: ActionManage, Scope: ScopeAll}
	PermProjectManageOwn  = PermissionRule{Resource: ResourceProject, Action: ActionManage, Scope: ScopeOwn}

	// Container permissions
	PermContainerReadAll     = PermissionRule{Resource: ResourceContainer, Action: ActionRead, Scope: ScopeAll}
	PermContainerReadTeam    = PermissionRule{Resource: ResourceContainer, Action: ActionRead, Scope: ScopeTeam}
	PermContainerCreateAll   = PermissionRule{Resource: ResourceContainer, Action: ActionCreate, Scope: ScopeAll}
	PermContainerCreateTeam  = PermissionRule{Resource: ResourceContainer, Action: ActionCreate, Scope: ScopeTeam}
	PermContainerCreateOwn   = PermissionRule{Resource: ResourceContainer, Action: ActionCreate, Scope: ScopeOwn}
	PermContainerUpdateAll   = PermissionRule{Resource: ResourceContainer, Action: ActionUpdate, Scope: ScopeAll}
	PermContainerUpdateTeam  = PermissionRule{Resource: ResourceContainer, Action: ActionUpdate, Scope: ScopeTeam}
	PermContainerDeleteAll   = PermissionRule{Resource: ResourceContainer, Action: ActionDelete, Scope: ScopeAll}
	PermContainerManageAll   = PermissionRule{Resource: ResourceContainer, Action: ActionManage, Scope: ScopeAll}
	PermContainerExecuteAll  = PermissionRule{Resource: ResourceContainer, Action: ActionExecute, Scope: ScopeAll}
	PermContainerExecuteTeam = PermissionRule{Resource: ResourceContainer, Action: ActionExecute, Scope: ScopeTeam}

	// Container Version permissions
	PermContainerVersionReadAll    = PermissionRule{Resource: ResourceContainerVersion, Action: ActionRead, Scope: ScopeAll}
	PermContainerVersionReadTeam   = PermissionRule{Resource: ResourceContainerVersion, Action: ActionRead, Scope: ScopeTeam}
	PermContainerVersionCreateAll  = PermissionRule{Resource: ResourceContainerVersion, Action: ActionCreate, Scope: ScopeAll}
	PermContainerVersionCreateTeam = PermissionRule{Resource: ResourceContainerVersion, Action: ActionCreate, Scope: ScopeTeam}
	PermContainerVersionUpdateAll  = PermissionRule{Resource: ResourceContainerVersion, Action: ActionUpdate, Scope: ScopeAll}
	PermContainerVersionUpdateTeam = PermissionRule{Resource: ResourceContainerVersion, Action: ActionUpdate, Scope: ScopeTeam}
	PermContainerVersionDeleteAll  = PermissionRule{Resource: ResourceContainerVersion, Action: ActionDelete, Scope: ScopeAll}
	PermContainerVersionManageAll  = PermissionRule{Resource: ResourceContainerVersion, Action: ActionManage, Scope: ScopeAll}
	PermContainerVersionUploadAll  = PermissionRule{Resource: ResourceContainerVersion, Action: ActionUpload, Scope: ScopeAll}
	PermContainerVersionUploadTeam = PermissionRule{Resource: ResourceContainerVersion, Action: ActionUpload, Scope: ScopeTeam}

	// Dataset permissions
	PermDatasetReadAll    = PermissionRule{Resource: ResourceDataset, Action: ActionRead, Scope: ScopeAll}
	PermDatasetReadTeam   = PermissionRule{Resource: ResourceDataset, Action: ActionRead, Scope: ScopeTeam}
	PermDatasetCreateAll  = PermissionRule{Resource: ResourceDataset, Action: ActionCreate, Scope: ScopeAll}
	PermDatasetCreateTeam = PermissionRule{Resource: ResourceDataset, Action: ActionCreate, Scope: ScopeTeam}
	PermDatasetCreateOwn  = PermissionRule{Resource: ResourceDataset, Action: ActionCreate, Scope: ScopeOwn}
	PermDatasetUpdateAll  = PermissionRule{Resource: ResourceDataset, Action: ActionUpdate, Scope: ScopeAll}
	PermDatasetUpdateTeam = PermissionRule{Resource: ResourceDataset, Action: ActionUpdate, Scope: ScopeTeam}
	PermDatasetDeleteAll  = PermissionRule{Resource: ResourceDataset, Action: ActionDelete, Scope: ScopeAll}
	PermDatasetManageAll  = PermissionRule{Resource: ResourceDataset, Action: ActionManage, Scope: ScopeAll}

	// Dataset Version permissions
	PermDatasetVersionReadAll      = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionRead, Scope: ScopeAll}
	PermDatasetVersionReadTeam     = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionRead, Scope: ScopeTeam}
	PermDatasetVersionCreateAll    = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionCreate, Scope: ScopeAll}
	PermDatasetVersionCreateTeam   = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionCreate, Scope: ScopeTeam}
	PermDatasetVersionUpdateAll    = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionUpdate, Scope: ScopeAll}
	PermDatasetVersionUpdateTeam   = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionUpdate, Scope: ScopeTeam}
	PermDatasetVersionDeleteAll    = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionDelete, Scope: ScopeAll}
	PermDatasetVersionManageAll    = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionManage, Scope: ScopeAll}
	PermDatasetVersionDownloadAll  = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionDownload, Scope: ScopeAll}
	PermDatasetVersionDownloadTeam = PermissionRule{Resource: ResourceDatasetVersion, Action: ActionDownload, Scope: ScopeTeam}

	// Label permissions
	PermLabelReadAll   = PermissionRule{Resource: ResourceLabel, Action: ActionRead, Scope: ScopeAll}
	PermLabelCreateAll = PermissionRule{Resource: ResourceLabel, Action: ActionCreate, Scope: ScopeAll}
	PermLabelCreateOwn = PermissionRule{Resource: ResourceLabel, Action: ActionCreate, Scope: ScopeOwn}
	PermLabelUpdateAll = PermissionRule{Resource: ResourceLabel, Action: ActionUpdate, Scope: ScopeAll}
	PermLabelDeleteAll = PermissionRule{Resource: ResourceLabel, Action: ActionDelete, Scope: ScopeAll}

	// Injection permissions
	PermInjectionReadProject     = PermissionRule{Resource: ResourceInjection, Action: ActionRead, Scope: ScopeProject}
	PermInjectionCreateProject   = PermissionRule{Resource: ResourceInjection, Action: ActionCreate, Scope: ScopeProject}
	PermInjectionUpdateProject   = PermissionRule{Resource: ResourceInjection, Action: ActionUpdate, Scope: ScopeProject}
	PermInjectionDeleteProject   = PermissionRule{Resource: ResourceInjection, Action: ActionDelete, Scope: ScopeProject}
	PermInjectionExecuteProject  = PermissionRule{Resource: ResourceInjection, Action: ActionExecute, Scope: ScopeProject}
	PermInjectionCloneProject    = PermissionRule{Resource: ResourceInjection, Action: ActionClone, Scope: ScopeProject}
	PermInjectionDownloadProject = PermissionRule{Resource: ResourceInjection, Action: ActionDownload, Scope: ScopeProject}

	// Execution permissions
	PermExecutionReadProject    = PermissionRule{Resource: ResourceExecution, Action: ActionRead, Scope: ScopeProject}
	PermExecutionCreateProject  = PermissionRule{Resource: ResourceExecution, Action: ActionCreate, Scope: ScopeProject}
	PermExecutionUpdateProject  = PermissionRule{Resource: ResourceExecution, Action: ActionUpdate, Scope: ScopeProject}
	PermExecutionDeleteProject  = PermissionRule{Resource: ResourceExecution, Action: ActionDelete, Scope: ScopeProject}
	PermExecutionExecuteProject = PermissionRule{Resource: ResourceExecution, Action: ActionExecute, Scope: ScopeProject}
	PermExecutionStopProject    = PermissionRule{Resource: ResourceExecution, Action: ActionStop, Scope: ScopeProject}

	// Task permissions
	PermTaskReadAll    = PermissionRule{Resource: ResourceTask, Action: ActionRead, Scope: ScopeAll}
	PermTaskCreateAll  = PermissionRule{Resource: ResourceTask, Action: ActionCreate, Scope: ScopeAll}
	PermTaskUpdateAll  = PermissionRule{Resource: ResourceTask, Action: ActionUpdate, Scope: ScopeAll}
	PermTaskDeleteAll  = PermissionRule{Resource: ResourceTask, Action: ActionDelete, Scope: ScopeAll}
	PermTaskExecuteAll = PermissionRule{Resource: ResourceTask, Action: ActionExecute, Scope: ScopeAll}
	PermTaskStopAll    = PermissionRule{Resource: ResourceTask, Action: ActionStop, Scope: ScopeAll}

	// Trace permissions
	PermTraceReadAll    = PermissionRule{Resource: ResourceTrace, Action: ActionRead, Scope: ScopeAll}
	PermTraceMonitorAll = PermissionRule{Resource: ResourceTrace, Action: ActionMonitor, Scope: ScopeAll}
)

// SystemRolePermissions defines the default permission rules for each system role
var SystemRolePermissions = map[RoleName][]PermissionRule{
	// System Roles
	RoleSuperAdmin: {}, // Super admin has unrestricted access to all resources

	RoleAdmin: {
		// System management
		PermSystemRead,
		PermSystemConfigure,
		PermSystemManage,
		PermAuditRead,
		PermAuditAudit,
		PermConfigurationRead,
		PermConfigurationUpdate,
		PermConfigurationConfigure,

		// User and permission management
		PermUserReadAll,
		PermUserCreateAll,
		PermUserUpdateAll,
		PermUserDeleteAll,
		PermUserAssignAll,
		PermRoleReadAll,
		PermRoleCreateAll,
		PermRoleUpdateAll,
		PermRoleDeleteAll,
		PermRoleGrantAll,
		PermRoleRevokeAll,
		PermPermissionReadAll,
		PermPermissionCreateAll,
		PermPermissionUpdateAll,
		PermPermissionDeleteAll,
		PermPermissionManageAll,

		// Team management
		PermTeamReadAll,
		PermTeamCreateAll,
		PermTeamUpdateAll,
		PermTeamDeleteAll,
		PermTeamManageAll,

		// Project management
		PermProjectReadAll,
		PermProjectUpdateAll,
		PermProjectDeleteAll,
		PermProjectManageAll,

		// Container management
		PermContainerReadAll,
		PermContainerCreateAll,
		PermContainerUpdateAll,
		PermContainerDeleteAll,
		PermContainerManageAll,
		PermContainerExecuteAll,
		PermContainerVersionReadAll,
		PermContainerVersionCreateAll,
		PermContainerVersionUpdateAll,
		PermContainerVersionDeleteAll,
		PermContainerVersionUploadAll,

		// Dataset management
		PermDatasetReadAll,
		PermDatasetCreateAll,
		PermDatasetUpdateAll,
		PermDatasetDeleteAll,
		PermDatasetManageAll,
		PermDatasetVersionReadAll,
		PermDatasetVersionCreateAll,
		PermDatasetVersionUpdateAll,
		PermDatasetVersionDeleteAll,
		PermDatasetVersionDownloadAll,

		// Label management
		PermLabelReadAll,
		PermLabelCreateAll,
		PermLabelUpdateAll,
		PermLabelDeleteAll,

		// Injection management
		PermInjectionReadProject,
		PermInjectionCreateProject,
		PermInjectionUpdateProject,
		PermInjectionDeleteProject,
		PermInjectionExecuteProject,
		PermInjectionCloneProject,
		PermInjectionDownloadProject,

		// Execution management
		PermExecutionReadProject,
		PermExecutionCreateProject,
		PermExecutionUpdateProject,
		PermExecutionDeleteProject,
		PermExecutionExecuteProject,
		PermExecutionStopProject,

		// Task management
		PermTaskReadAll,
		PermTaskCreateAll,
		PermTaskUpdateAll,
		PermTaskDeleteAll,
		PermTaskExecuteAll,
		PermTaskStopAll,

		// Trace management
		PermTraceReadAll,
		PermTraceMonitorAll,
	},

	// Regular User Role - basic role with minimal permissions
	// Users can read team resources if they are team members
	RoleUser: {
		// Project permissions
		PermProjectCreateOwn,
		PermProjectReadOwn,

		// Container permissions
		PermContainerCreateOwn,

		// Dataset permissions
		PermDatasetCreateOwn,

		// Label permissions
		PermLabelCreateOwn,
		PermLabelReadAll,
	},

	// Container Roles
	RoleContainerAdmin: {
		PermContainerReadAll,
		PermContainerCreateAll,
		PermContainerUpdateAll,
		PermContainerDeleteAll,
		PermContainerManageAll,
		PermContainerExecuteAll,
		PermContainerVersionReadAll,
		PermContainerVersionCreateAll,
		PermContainerVersionUpdateAll,
		PermContainerVersionDeleteAll,
		PermContainerVersionManageAll,
		PermContainerVersionUploadAll,
	},
	RoleContainerDeveloper: {
		PermContainerReadTeam,
		PermContainerCreateTeam,
		PermContainerUpdateTeam,
		PermContainerExecuteTeam,
		PermContainerVersionReadTeam,
		PermContainerVersionCreateTeam,
		PermContainerVersionUpdateTeam,
		PermContainerVersionUploadTeam,
	},
	RoleContainerViewer: {
		PermContainerReadAll,
		PermContainerVersionReadAll,
	},

	// Dataset Roles
	RoleDatasetAdmin: {
		PermDatasetReadAll,
		PermDatasetCreateAll,
		PermDatasetUpdateAll,
		PermDatasetDeleteAll,
		PermDatasetManageAll,
		PermDatasetVersionReadAll,
		PermDatasetVersionCreateAll,
		PermDatasetVersionUpdateAll,
		PermDatasetVersionDeleteAll,
		PermDatasetVersionManageAll,
		PermDatasetVersionDownloadAll,
	},
	RoleDatasetDeveloper: {
		PermDatasetReadTeam,
		PermDatasetCreateTeam,
		PermDatasetUpdateTeam,
		PermDatasetVersionReadTeam,
		PermDatasetVersionCreateTeam,
		PermDatasetVersionUpdateTeam,
		PermDatasetVersionDownloadTeam,
	},
	RoleDatasetViewer: {
		PermDatasetReadAll,
		PermDatasetVersionReadAll,
		PermDatasetVersionDownloadAll,
	},

	// Team Roles
	RoleTeamAdmin: {
		PermTeamReadAll,
		PermTeamCreateAll,
		PermTeamUpdateAll,
		PermTeamDeleteAll,
		PermTeamManageAll,
	},
	RoleTeamMember: {
		PermTeamReadTeam,
		PermTeamUpdateTeam,
	},
	RoleTeamViewer: {
		PermTeamReadTeam,
	},

	// Project Roles
	RoleProjectAdmin: {
		PermProjectReadOwn,
		PermProjectUpdateOwn,
		PermProjectDeleteOwn,
		PermProjectManageOwn,
		// Injection permissions
		PermInjectionReadProject,
		PermInjectionExecuteProject,
		// Execution permissions
		PermExecutionReadProject,
		PermExecutionExecuteProject,
	},
	RoleProjectAlgoDeveloper: {
		PermProjectReadOwn,
		// Execution permissions
		PermExecutionReadProject,
		PermExecutionExecuteProject,
	},
	RoleProjectDataDeveloper: {
		PermProjectReadOwn,
		// Injection permissions
		PermInjectionReadProject,
		PermInjectionExecuteProject,
	},
	RoleProjectViewer: {
		PermProjectReadOwn,
		PermInjectionReadProject,
		PermExecutionReadProject,
	},
}
