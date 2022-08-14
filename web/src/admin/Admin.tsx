import React, {FunctionComponent, useEffect, useState} from "react";
import {ApiDescription, ApiType} from "../types";
import jsonServerProvider from "ra-data-json-server";
import {CustomTheme} from "../theme/customTheme";
import {Datagrid, DeleteButton, EditButton, List, Resource, TextField, Admin as ReactAdmin} from "react-admin";
import {JsrfCreate, JsrfEdit} from "../jsrf/JsrfEdit";
import {flattenSchema} from "../schema/json-schema";

type AdminProps = {
    server: string
}

export const Admin = ({server}: AdminProps) => {

    let data = jsonServerProvider(server);
    const [apiDescription, setApiDescription] = useState<ApiDescription>({types: []});

    useEffect(() => {
        async function getApiDescription() {
            const response = await fetch(server + "/describe");
            const body = await response.json();
            setApiDescription(body)
        }

        getApiDescription()
    }, [])

    const resources = apiDescription.types.map((t) => typeToResource(t));

    if (resources.length === 0) {
        return (<div></div>)
    }

    return (
        <ReactAdmin disableTelemetry={true} dataProvider={data} theme={CustomTheme}>
            {resources}
        </ReactAdmin>
    );
}


function typeToResource(t: ApiType) {
    return (<Resource key={t.name}
                      name={t.name}
                      list={schemaToList(t.schema)}
                      create={<JsrfCreate type={t}/>}
                      edit={<JsrfEdit type={t}/>}
    />)
}


function schemaToList(schema: string) {
    //TODO test with nested prop
    //TODO may prop type to widget type
    //TODO array types (and parsing of)
    const fields = flattenSchema(schema).map(p => <TextField key={p.path} source={p.path}/>);

    return <List>
        <Datagrid>
            {fields}
            <EditButton/>
            <DeleteButton/>
        </Datagrid>
    </List>
}
