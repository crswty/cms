type JsonSchemaSimple = {
    type: "string" | "number" | "integer" | "boolean"
}

type JsonSchemaObject = {
    type: "object"
    properties: JsonSchemaPropertySet
}

type JsonSchemaArray = {
    type: "array"
    items: JsonSchemaProperty
}

type JsonSchemaProperty = JsonSchemaSimple | JsonSchemaObject | JsonSchemaArray

type JsonSchemaPropertySet = {
    [key: string]: JsonSchemaProperty
}
type JsonSchema = {
    properties: JsonSchemaPropertySet
}

export type Property  = {
    path: string
    type: string
}


function getProps(path: string, props: JsonSchemaPropertySet) {
    const flatProps: Property[] = []
    for (let [key, prop] of Object.entries(props)) {

        switch (prop.type) {
            case "object": {
                flatProps.push(...getProps(path + key + ".", prop.properties))
                break;
            }
            default: {
                flatProps.push({path: path + key, type: prop.type} as Property)
            }

        }
    }
    return flatProps;
}

export const flattenSchema = (schema: string): Property[] => {
    const s = JSON.parse(schema) as JsonSchema;
    return getProps("", s.properties);
}