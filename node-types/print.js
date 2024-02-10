function filterNamed(list) {
  return list.filter(e => e.named).map(e => e.type);
}

const fs = require("fs")
const types = JSON.parse(fs.readFileSync("node-types.json", "utf8"))

const concreteType = []
for (const t of types) {
  if (t.fields || t.children) {
    let fields = undefined
    if (t.fields) {
      fields = {}
      for (const f in t.fields) {
        // fields[f] = {
        //   ...t.fields[f],
        //   types: filterNamed(t.fields[f].types)
        // }
        const types = filterNamed(t.fields[f].types)
        if (types.length === 1) {
          fields[f] = types[0]
        } else {
          fields[f] = types
        }
      }
    }

    let children = undefined
    if (t.children) {
      // children = {
      //   ...t.children,
      //   types: filterNamed(t.children.types)
      // }
      const types = filterNamed(t.children.types)
      if (types.length === 1) {
        children = types[0]
      } else {
        children = types
      }
    }

    concreteType.push({
      type: t.type,
      fields: fields,
      children: children,
    })
  }
}

for (const t of types) {
  if (t.subtypes) {
    console.log(`======= ${t.type} =======`)
    for (const subType of t.subtypes) {
      console.log(` - ${subType.type}`)
    }
  }
}

console.log(JSON.stringify(concreteType, undefined, 2))
