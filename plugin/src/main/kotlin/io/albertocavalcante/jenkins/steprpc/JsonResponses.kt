package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.Message
import io.albertocavalcante.jenkins.steprpc.v1.Error
import io.albertocavalcante.jenkins.steprpc.v1.ErrorResponse
import java.nio.charset.StandardCharsets
import net.sf.json.JSONObject
import org.kohsuke.stapler.HttpResponse
import org.kohsuke.stapler.StaplerRequest2
import org.kohsuke.stapler.StaplerResponse2

fun jsonResponse(payload: Map<String, Any?>, statusCode: Int = 200): HttpResponse {
    val body = JSONObject.fromObject(payload).toString().toByteArray(StandardCharsets.UTF_8)
    return responseWithBody(statusCode, body)
}

fun jsonResponse(payload: Message, statusCode: Int = 200): HttpResponse {
    val body = protoToJson(payload).toByteArray(StandardCharsets.UTF_8)
    return responseWithBody(statusCode, body)
}

fun errorResponse(statusCode: Int, code: String, message: String, details: Map<String, String> = emptyMap()): HttpResponse {
    val error = Error.newBuilder()
        .setCode(code)
        .setMessage(message)
        .putAllDetails(details)
        .build()

    return jsonResponse(
        payload = ErrorResponse.newBuilder().setError(error).build(),
        statusCode = statusCode,
    )
}

private fun responseWithBody(statusCode: Int, body: ByteArray): HttpResponse {
    return object : HttpResponse {
        override fun generateResponse(req: StaplerRequest2, rsp: StaplerResponse2, node: Any?) {
            rsp.status = statusCode
            rsp.contentType = "application/json; charset=UTF-8"
            rsp.setContentLength(body.size)
            rsp.outputStream.write(body)
        }
    }
}
