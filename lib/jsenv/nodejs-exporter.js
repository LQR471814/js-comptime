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
