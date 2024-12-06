import { useState, useEffect } from "react"
import { getServiceListApi, getNamespacesApi, getServiceEndpointNameApi } from "src/core/api/service"
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange, toNearestSecond } from 'src/core/store/reducers/timeRangeReducer'
import { Select } from "antd"
import { getStep } from 'src/core/utils/step'


const TableFilter = ({ getTableData }) => {
    const [serviceNameOptions, setServiceNameOptions] = useState([])
    const [endpointNameOptions, setEndpointNameOptions] = useState([])
    const [namespaceOptions, setNamespaceOptions] = useState([])
    const [serachServiceName, setSerachServiceName] = useState()
    const [serachEndpointName, setSerachEndpointName] = useState()
    const [serachNamespace, setSerachNamespace] = useState()

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
        try {
            const data = await getNamespacesApi();
            const mapData = (data?.items || []).map(item => ({
                value: item.metadata.name,
                label: item.metadata.name,
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

    useEffect(() => {
        const fetchData = async () => {
            await Promise.all([getServiceNameOptions(), getNamespaceOptions()]);
        };

        fetchData();
    }, []);

    useEffect(() => {
        if (serachServiceName?.length > 0) {
            getEndpointNameOptions(serachServiceName);
        }
    }, [serachServiceName]);

    useEffect(() => {
        getTableData({ serachServiceName, serachEndpointName, serachNamespace })
    }, [serachServiceName, serachEndpointName, serachNamespace])

    return (<>
        <div className="p-2 my-2 flex flex-row w-full">
            <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
                <span className="text-nowrap">命名空间：</span>
                <Select
                    mode='multiple'
                    id="namespace"
                    placeholder="请选择"
                    className='w-full'
                    value={serachNamespace}
                    onChange={(e) => setSerachNamespace(e)}
                    options={namespaceOptions}
                    maxTagCount={2}
                    maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
                    allowClear
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
                    onChange={(e) => setSerachServiceName(e)}
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

export default TableFilter