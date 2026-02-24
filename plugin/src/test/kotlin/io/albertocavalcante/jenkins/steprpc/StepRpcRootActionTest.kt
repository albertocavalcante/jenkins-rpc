package io.albertocavalcante.jenkins.steprpc

import org.junit.jupiter.api.Test
import kotlin.test.assertNotNull

class StepRpcRootActionTest {
    @Test
    fun `root action can be instantiated`() {
        val action = StepRpcRootAction()
        assertNotNull(action)
        assertNotNull(action.getV1())
    }
}
