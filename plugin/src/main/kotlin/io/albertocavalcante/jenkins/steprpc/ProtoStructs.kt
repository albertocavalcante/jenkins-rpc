package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.Struct
import com.google.protobuf.Value

fun structToAnyMap(struct: Struct): Map<String, Any?> {
    return struct.fieldsMap.mapValues { (_, value) -> valueToAny(value) }
}

private fun valueToAny(value: Value): Any? {
    return when (value.kindCase) {
        Value.KindCase.NULL_VALUE -> null
        Value.KindCase.NUMBER_VALUE -> value.numberValue
        Value.KindCase.STRING_VALUE -> value.stringValue
        Value.KindCase.BOOL_VALUE -> value.boolValue
        Value.KindCase.STRUCT_VALUE -> structToAnyMap(value.structValue)
        Value.KindCase.LIST_VALUE -> value.listValue.valuesList.map { valueToAny(it) }
        Value.KindCase.KIND_NOT_SET -> null
    }
}
