import {flattenSchema} from "./json-schema";


describe('json-schema', () => {

    it('get flattened properties', () => {

        const properties = flattenSchema(JSON.stringify(testSchema));
        expect(properties).toHaveLength(3)
        expect(properties).toContainEqual({path: "id", type: "string"})
        expect(properties).toContainEqual({path: "name", type: "string"})
        expect(properties).toContainEqual({path: "parent.nested1", type: "string"})
    })


})

const testSchema = {
    "$id": "http://example.com/schema/my-test-schema",
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "type": "object",
    "required": ["id", "name"],
    "properties": {
        "id": {
            "type": "string"
        },
        "name": {
            "type": "string"
        },
        "parent": {
            "type": "object",
            "properties": {
                "nested1": {
                    "type": "string"
                }
            }
        }
    }
}