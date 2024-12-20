import { useLogsContext } from 'src/core/contexts/LogsContext';
import LogValueTag from './LogValueTag';
import LogKeyTag from './LogKeyTag';
import { useMemo } from 'react';

const LogItemFold = ({ tags }) => {
  const { tableInfo, displayFields } = useLogsContext();
  //由tableName和type组成的唯一标识
  const tableId = `${tableInfo.tableName}_${tableInfo.type}`
  // 计算过滤后的 tags
  const filteredTags = useMemo(() => {
    return Object.entries(tags).filter(([key, value]) => {
      return (
        displayFields[tableId]?.includes(key) &&  // 检查是否在显示字段中
        value !== '' &&  // 确保值不为空
        key !== (tableInfo?.timeField || 'timestamp') &&  // 排除时间字段
        typeof value !== 'object'  // 确保值不是对象
      );
    });
  }, [tags, displayFields, tableInfo?.timeField]);

  return (
    <>
      {/* 渲染 tags */}
      <div className="text-ellipsis text-wrap flex" style={{ display: '-webkit-box' }}>
        {filteredTags.map(([key, value]) => (
          <LogValueTag key={key} objKey={key} value={String(value)} />
        ))}
      </div>
    </>
  );
};

export default LogItemFold;
