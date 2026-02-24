package io.albertocavalcante.jenkins.steprpc

import jenkins.model.Jenkins
import org.junit.ClassRule
import org.junit.Test
import org.jvnet.hudson.test.JenkinsRule
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class StepRpcPermissionsTest {
    companion object {
        @ClassRule
        @JvmField
        val j = JenkinsRule()
    }

    @Test
    fun `INVOKE permission exists`() {
        assertNotNull(StepRpcPermissions.INVOKE)
    }

    @Test
    fun `INVOKE permission belongs to GROUP`() {
        assertEquals(StepRpcPermissions.GROUP, StepRpcPermissions.INVOKE.group)
    }

    @Test
    fun `INVOKE permission is named Invoke`() {
        assertEquals("Invoke", StepRpcPermissions.INVOKE.name)
    }

    @Test
    fun `INVOKE permission is implied by ADMINISTER`() {
        assertEquals(Jenkins.ADMINISTER, StepRpcPermissions.INVOKE.impliedBy)
    }

    @Test
    fun `GROUP is bound to StepRpcRootAction`() {
        assertEquals(StepRpcRootAction::class.java, StepRpcPermissions.GROUP.owner)
    }
}
