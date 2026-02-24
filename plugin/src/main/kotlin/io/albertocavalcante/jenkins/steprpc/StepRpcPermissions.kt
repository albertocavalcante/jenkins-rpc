package io.albertocavalcante.jenkins.steprpc

import hudson.security.Permission
import hudson.security.PermissionGroup
import hudson.security.PermissionScope
import jenkins.model.Jenkins

object StepRpcPermissions {
    @JvmField
    val GROUP: PermissionGroup = PermissionGroup(
        StepRpcRootAction::class.java,
        Messages._StepRpc_Permissions_Title(),
    )

    @JvmField
    val INVOKE: Permission = Permission(
        GROUP,
        "Invoke",
        Messages._StepRpc_Permission_Invoke_Description(),
        Jenkins.ADMINISTER,
        PermissionScope.JENKINS,
    )
}
