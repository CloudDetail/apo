import React, { forwardRef, useImperativeHandle } from "react";
import { Button, Form, Input, Select } from "antd";
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import { useEffect } from "react";

const typeOptions = [
    { value: "Int8", label: "Int8" },
    { value: "Int16", label: "Int16" },
    { value: "Int32", label: "Int32" },
    { value: "Int64", label: "Int64" },
    { value: "Int128", label: "Int128" },
    { value: "Int256", label: "Int256" },
    { value: "UInt8", label: "UInt8" },
    { value: "UInt16", label: "UInt16" },
    { value: "UInt32", label: "UInt32" },
    { value: "UInt64", label: "UInt64" },
    { value: "UInt128", label: "UInt128" },
    { value: "UInt256", label: "UInt256" },
    { value: "Float32", label: "Float32" },
    { value: "Float64", label: "Float64" },
    { value: "Date", label: "Date" },
    { value: "Date32", label: "Date32" },
    { value: "DateTime", label: "DateTime" },
    { value: "DateTime64", label: "DateTime64" },
    { value: "String", label: "String" },
    { value: "FixedString(N)", label: "FixedString(N)" },
    { value: "Bool", label: "Bool" },
];

const LogStructRuleFormList = forwardRef(({ object, setObject }, ref) => {
    const [form] = Form.useForm();
    useImperativeHandle(ref, () => {
        return form
    }, [form])

    useEffect(() => {
        form.resetFields();
    }, [object]);

    useEffect(() => {
        setObject([
            {
                name: "",
                type: "String"
            }
        ])
    }, [])

    const removeStructuring = (index) => {
        setObject(object.filter((_, i) => i !== index))
    }

    const updateObject = () => {
        const oldStructuringObject = []
        Object.keys(form.getFieldsValue()).forEach(key => {
            const match = key.match(/^(\w+)_Type$/);
            if (match) {
                const field = match[1];
                oldStructuringObject.push({
                    type: form.getFieldsValue()[`${field}_Type`],
                    name: form.getFieldsValue()[`${field}_Name`]
                });
            }
        });

        setObject([...oldStructuringObject])
    }

    return (
        <>
            <Form className="w-full" form={form}>
                {object?.map((item, index) => (
                    <div className="flex w-10/12 mb-8" key={index}>
                        <Form.Item
                            className="w-1/2 mr-3"
                            // name={key + "_" + "Name"}
                            name={index + "_" + "Name"}
                            initialValue={item["name"]}
                            rules={[
                                {
                                    required: true,
                                    message: "请填写字段名"
                                }
                            ]}
                        >
                            <Input placeholder="字段名" onBlur={updateObject} />
                        </Form.Item>
                        <Form.Item
                            className="w-1/6"
                            // name={key + "_" + "Type"}
                            name={index + "_" + "Type"}
                            rules={[
                                {
                                    required: true,
                                    message: "请选择类型",
                                },
                            ]}
                            initialValue={item["type"]}
                        >
                            <Select options={typeOptions} onChange={updateObject} />
                        </Form.Item>
                        <IoIosRemoveCircleOutline
                            size={20}
                            className="mt-1 cursor-pointer ml-5"
                            onClick={() => {
                                removeStructuring(index)
                            }}
                        />
                    </div>
                ))}
            </Form>
        </>
    );
});

export default LogStructRuleFormList;
