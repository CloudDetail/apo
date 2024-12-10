import { useState, useEffect, forwardRef, useImperativeHandle } from "react"
import { getServiceListApi, getNamespacesApi, getServiceEndpointNameApi } from "src/core/api/service"
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange, toNearestSecond } from 'src/core/store/reducers/timeRangeReducer'
import { Select } from "antd"
import { getStep } from 'src/core/utils/step'


export const TableFilter = (props) => {
    const {
        setServiceName,
        setEndpoint,
        setNamespace
    } = props

    const [serviceNameOptions, setServiceNameOptions] = useState([])
    const [endpointNameOptions, setEndpointNameOptions] = useState([])
    const [namespaceOptions, setNamespaceOptions] = useState([])
    const [serachServiceName, setSerachServiceName] = useState(null)
    const [serachEndpointName, setSerachEndpointName] = useState(null)
    const [serachNamespace, setSerachNamespace] = useState(null)
    const [prevSearchServiceName, setPrevSearchServiceName] = useState(null)

    const { startTime, endTime } = useSelector(selectSecondsTimeRange)

    const getServiceNameOptions = async () => {
        try {
            const params = { startTime, endTime };
            const data = await getServiceListApi(params);
            setServiceNameOptions(data.map(value => ({ value, label: value })));
        } catch (error) {
            console.error('获取服务列表失败:', error);
        }
    };

    const getNamespaceOptions = async () => {
        const params = { startTime, endTime }
        try {
            const data = await getNamespacesApi(params);
            const mapData = (data?.namespaceList || []).map(namespace => ({
                value: namespace,
                label: namespace,
            }));
            setNamespaceOptions(mapData);
        } catch (error) {
            console.log('获取命名空间失败:', error);
        }
    };

    const getEndpointNameOptions = async (serviceNameList) => {
        setEndpointNameOptions([]);
        try {
            const newEndpointNameOptions = await Promise.all(serviceNameList.map(async (element) => {
                const params = {
                    startTime,
                    endTime,
                    step: getStep(startTime, endTime),
                    serviceName: element,
                    sortRule: 1,
                };
                const data = await getServiceEndpointNameApi(params);
                return {
                    label: <span>{element}</span>,
                    title: element,
                    options: data.map(item => ({
                        label: <span>{item?.endpoint}</span>,
                        value: item?.endpoint,
                    })),
                };
            }));

            setEndpointNameOptions(newEndpointNameOptions);
        } catch (error) {
            console.error('获取 endpoint 失败:', error);
        }
    };

    const onChangeNamespace = (event) => {
        setSerachNamespace(event)
    }

    const onChangeServiceName = (event) => {
        setPrevSearchServiceName(serachServiceName)
        setSerachServiceName(event)
    }

    const removeEndpointNames = () => {
        if (prevSearchServiceName?.length > serachServiceName?.length) {
            // 找出需要移除的服务名称
            const removeServiceNameSet = new Set(prevSearchServiceName.filter(item => !serachServiceName.includes(item)));

            // 找出需要移除的端点值
            const removeEndpoint = endpointNameOptions
                .flatMap(item => removeServiceNameSet.has(item.title) ? item.options : [])
                .map(item => item.value);

            // 从 serachEndpointName 中移除相关端点
            const removedSearchedName = serachEndpointName?.filter(item => !removeEndpoint?.includes(item));

            // 更新状态
            setSerachEndpointName(removedSearchedName);
        }

        // 获取最新的端点名称选项
        getEndpointNameOptions(serachServiceName);
    }

    useEffect(() => {
        const fetchData = async () => {
            await Promise.all([getServiceNameOptions(), getNamespaceOptions()]);
        };

        fetchData();
    }, []);

    useEffect(() => {
        removeEndpointNames()
    }, [serachServiceName]);

    useEffect(() => {
        setServiceName(serachServiceName)
        setEndpoint(serachEndpointName)
        setNamespace(serachNamespace)
    }, [serachServiceName, serachEndpointName, serachNamespace])

    return (<>
        <div className="p-2 my-2 flex flex-row w-full">
            <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
                <span className="text-nowrap">命名空间：</span>
                <Select
                    mode='multiple'
                    allowClear
                    id="namespace"
                    className='w-full'
                    placeholder="请选择"
                    value={serachNamespace}
                    onChange={onChangeNamespace}
                    options={namespaceOptions}
                    maxTagCount={2}
                    maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
                />
            </div>
            <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
                <span className="text-nowrap">服务名：</span>
                <Select
                    mode='multiple'
                    allowClear
                    className='w-full'
                    id="serviceName"
                    placeholder="请选择"
                    value={serachServiceName}
                    onChange={onChangeServiceName}
                    options={serviceNameOptions}
                    popupMatchSelectWidth={false}
                    maxTagCount={2}
                    maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
                />
            </div>
            <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
                <span className="text-nowrap">服务端点：</span>
                <Select
                    mode='multiple'
                    id="endpointName"
                    placeholder="请选择"
                    className='w-full'
                    value={serachEndpointName}
                    popupMatchSelectWidth={false}
                    onChange={(e) => setSerachEndpointName(e)}
                    options={endpointNameOptions}
                    maxTagCount={2}
                    maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
                    allowClear
                />
            </div>
            <div>{/* <ThresholdCofigModal /> */}</div>
        </div>
    </>)
}