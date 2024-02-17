let __jscomptime_export_value
{
    // wrapped in block to avoid polluting global scope
    const sock = require("node:dgram")
        .createSocket("udp4");
    const port = parseInt(process.env.JSCOMPTIME_PORT)
    function serializeValue(value) {
        const typeofStr = typeof value
        switch (typeofStr) {
            case "undefined":
            case "null":
                return typeofStr
            case "boolean":
            case "number":
                return value.toString()
            case "string":
                // JSON.stringify is used to escape the string.
                return JSON.stringify(value)
            case "symbol":
                // JSON.stringify is used to escape the string.
                return `Symbol("${JSON.stringify(value.description)}")`
            case "bigint":
                return `BigInt(${value.toString()})`
        }
    }
    __jscomptime_export_value = function(id, value) {
        const fullText = id + "|" + serializeValue(value) + "\0"
        let cursor = 0
        while (true) {
            const chunk = fullText.slice(cursor, cursor + 512)
            if (chunk === "") {
                break
            }
            sock.send(chunk, port, "127.0.0.1")
            cursor += 512
        }
    }
}
const comptimeVar = 24
function comptimeFunction(a, b) {
  return n ** 2
}
const comptimeKey = "foo"
__jscomptime_export_value(0, comptimeVar)
__jscomptime_export_value(1, [comptimeVar, 42])
__jscomptime_export_value(2, comptimeKey)
__jscomptime_export_value(3, comptimeVar)
__jscomptime_export_value(4, { comptimeVar, [comptimeKey]: comptimeVar })
__jscomptime_export_value(5, comptimeVar)
__jscomptime_export_value(6, comptimeKey)
__jscomptime_export_value(7, comptimeVar)
__jscomptime_export_value(8, comptimeVar)
__jscomptime_export_value(9, comptimeVar)
__jscomptime_export_value(10, comptimeKey)
__jscomptime_export_value(11, 1)
__jscomptime_export_value(12, comptimeVar)
__jscomptime_export_value(13, comptimeVar)
__jscomptime_export_value(14, comptimeKey)
__jscomptime_export_value(15, 23)
__jscomptime_export_value(16, [2, 3, 4])
__jscomptime_export_value(17, (true))
__jscomptime_export_value(18, (!false))
__jscomptime_export_value(19, (true))
