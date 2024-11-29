import React, { forwardRef, useImperativeHandle, useState } from "react";
import { Button, Form, Input, Select } from "antd";
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useEffect } from "react";
import TextArea from 'antd/es/input/TextArea'

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

const LogStructRuleFormList = forwardRef(({ jsonRule, fForm }, ref) => {
    const [structuringObject, setStructuringObject] = useState([])
    const [form] = Form.useForm();
    useImperativeHandle(ref, () => ({
        form,
        setStructuringObject,
    }), [form])

    const parseJsonRule = (jsonRule) => {
        const jsonForParse = jsonRule.trim(); // 去掉前后空格
        if (!jsonForParse) {
            setStructuringObject([
                {
                    name: "",
                    type: "String"
                }
            ]); // 如果为空，清空列表
            fForm.setFields([
                {
                    name: "structuredRule",
                    errors: []
                }
            ])
            return;
        }
        let parsedJson = null;
        try {
            parsedJson = JSON.parse(jsonForParse);
            fForm.setFields([
                {
                    name: "structuredRule",
                    errors: []
                }
            ])
        } catch (error) {
            console.error("解析失败:", error);
            fForm.setFields([
                {
                    name: "structuredRule",
                    errors: ["json解析失败，请检查格式是否正确"]
                }
            ])
            setStructuringObject([
                {
                    name: "",
                    type: "String"
                }
            ]);
            return;
        }

        if (typeof parsedJson === "object" && parsedJson !== null && JSON.stringify(parsedJson) !== '{}') {
            const result = Object.entries(parsedJson).map(([key, value]) => {
                let type;
                if (typeof value === 'string') {
                    type = 'String';
                } else if (typeof value === 'number') {
                    type = Number.isInteger(value) ? 'Int64' : 'Float64';
                } else if (typeof value === 'boolean') {
                    type = 'Bool';
                } else {
                    type = 'String';
                }
                return { name: key, type };
            });
            setStructuringObject(result)
        } else {
            fForm.setFields([
                {
                    name: "structuredRule",
                    errors: ["json解析失败，请检查格式是否正确"]
                }
            ])
        }
    };
    const addStructConf = () => {
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

        setStructuringObject([...oldStructuringObject, {
            name: "",
            type: "String"
        }])
    }

    useEffect(() => {
        form.resetFields();
    }, [structuringObject]);

    useEffect(() => {
        setStructuringObject([
            {
                name: "",
                type: "String"
            }
        ])
    }, [])

    useEffect(() => {
        parseJsonRule(jsonRule)
    }, [jsonRule])

    const removeStructuring = (index) => {
        setStructuringObject(structuringObject.filter((_, i) => i !== index))
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

        setStructuringObject([...oldStructuringObject])
    }

    return (
        <>
            <div className='flex items-center mt-2 mb-2 w-full justify-start'>
                <span className="text-md text-gray-400">日志解析规则</span>
                <div className='flex items-center'>
                    <IoMdAddCircleOutline
                        size={20}
                        className="cursor-pointer ml-2"
                        onClick={addStructConf}
                    />
                </div>
            </div>
            <div className='flex flex-col items-start w-full'>
                <Form className="w-full" form={form}>
                    {structuringObject?.map((item, index) => (
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
            </div>
        </>
    );
});

export default LogStructRuleFormList;
