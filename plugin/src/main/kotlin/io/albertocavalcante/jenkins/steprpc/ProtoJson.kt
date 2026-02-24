package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.Message
import com.google.protobuf.Message.Builder
import com.google.protobuf.util.JsonFormat

private val protoJsonPrinter = JsonFormat.printer().omittingInsignificantWhitespace()
private val protoJsonParser = JsonFormat.parser()

fun protoToJson(message: Message): String = protoJsonPrinter.print(message)

fun mergeJsonIntoBuilder(json: String, builder: Builder) {
    protoJsonParser.merge(json, builder)
}
