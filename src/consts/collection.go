package consts

// SystemRoleDisplayNames maps system role names to their display names
var SystemRoleDisplayNames = map[RoleName]string{
	RoleSuperAdmin:         "Super Admin",
	RoleAdmin:              "Admin",
	RoleContainerAdmin:     "Container Admin",
	RoleContainerDeveloper: "Container Developer",
	RoleContainerViewer:    "Container Viewer",
	RoleProjectAdmin:       "Project Admin",
	RoleProjectDeveloper:   "Project Developer",
	RoleProjectViewer:      "Project Viewer",
}

// SystemRolePermissions defines the default permissions for each system role
var SystemRolePermissions = map[RoleName][]PermissionName{
	RoleSuperAdmin: {},
	RoleAdmin: {
		PermissionReadProject, PermissionWriteProject, PermissionDeleteProject, PermissionManageProject,
		PermissionReadDataset, PermissionWriteDataset, PermissionDeleteDataset, PermissionManageDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionDeleteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadContainer, PermissionWriteContainer, PermissionDeleteContainer, PermissionManageContainer,
		PermissionReadTask, PermissionWriteTask, PermissionDeleteTask, PermissionExecuteTask,
		PermissionReadRole,
		PermissionReadPermission,
	},
	RoleContainerAdmin: {
		PermissionReadContainer, PermissionWriteContainer, PermissionDeleteContainer, PermissionManageContainer,
		PermissionReadContainerVersion, PermissionWriteContainerVersion, PermissionDeleteContainerVersion, PermissionManageContainerVersion,
	},
	RoleContainerDeveloper: {
		PermissionReadContainer, PermissionWriteContainer,
		PermissionReadContainerVersion, PermissionWriteContainerVersion,
	},
	RoleContainerViewer: {
		PermissionReadContainer,
		PermissionReadContainerVersion,
	},
	RoleProjectAdmin: {
		PermissionReadProject, PermissionWriteProject, PermissionDeleteProject, PermissionManageProject,
		PermissionReadDataset, PermissionWriteDataset, PermissionDeleteDataset, PermissionManageDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionDeleteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadTask, PermissionWriteTask, PermissionExecuteTask,
	},
	RoleProjectDeveloper: {
		PermissionReadProject,
		PermissionReadDataset, PermissionWriteDataset,
		PermissionReadFaultInjection, PermissionWriteFaultInjection, PermissionExecuteFaultInjection,
		PermissionReadTask, PermissionWriteTask, PermissionExecuteTask,
	},
	RoleProjectViewer: {
		PermissionReadProject,
		PermissionReadDataset,
		PermissionReadFaultInjection,
		PermissionReadTask,
	},
}
