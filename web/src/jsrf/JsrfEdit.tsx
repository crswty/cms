import {useNavigate, useParams} from "react-router-dom";
import {EditBase, CreateBase, useCreate, useGetOne, useUpdate} from "react-admin";
import {MuiForm5 as Form} from "@rjsf/material-ui";
import React, {FunctionComponent} from "react";
import {ApiType} from "../types";
import './JsrfEdit.css'

type JsrfEditProps = {
    type: ApiType
}

export const JsrfEdit: FunctionComponent<JsrfEditProps> = ({type}) => {
    const {id} = useParams();
    const [update] = useUpdate();
    const navigate = useNavigate();

    const {data} = useGetOne(type.name, {id});

    const onSubmit = async (formData: any) => {
        await update(type.name,
            {id: formData[type.id], data: formData},
            {
                onSuccess: () => {
                    navigate("/" + type.name)
                }
            })
    }

    const schema = JSON.parse(type.schema);
    delete schema.$schema

    return (
        <EditBase>
            <div>
                <Form formData={data} schema={schema}
                      onSubmit={(a) => onSubmit(a.formData)}/>
            </div>
        </EditBase>
    );
}

export const JsrfCreate: FunctionComponent<JsrfEditProps> = ({type}) => {
    const [create] = useCreate();
    const navigate = useNavigate();

    const onSubmit = async (formData: any) => {
        await create(type.name,
            {data: formData},
            {
                onSuccess: () => {
                    navigate("/" + type.name)
                }
            })
    }

    const schema = JSON.parse(type.schema);
    delete schema.$schema

    return (
        <CreateBase>
            <div>
                <Form schema={schema}
                      onSubmit={(a) => onSubmit(a.formData)}/>
            </div>
        </CreateBase>
    );
}