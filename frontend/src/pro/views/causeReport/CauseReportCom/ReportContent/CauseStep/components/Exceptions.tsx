// import React, { useState } from 'react';

// function ExceptionsTable(props: any) {
//     const { exceptions,spanName } = props;
//     const [openExceptionsList, setOpenExceptionsList] = useState<number[]>([]);

//     const clickCollapsableSection = (index: number) => {
//         if (openExceptionsList.includes(index)) {
//             setOpenExceptionsList(openExceptionsList.filter(item => item !== index));
//         } else {
//             setOpenExceptionsList([...openExceptionsList, index]);
//         }
//     }

//     return (
//         <>
//             {exceptions.map((item: any, index: number) => (
//                 <Collapse className="exception-collapse" collapsible label={item.message} key={index} isOpen={openExceptionsList.includes(index)} onToggle={() => clickCollapsableSection(index)}>
//                     <div className="row">
//                         <div className={style.exceptionLabel + " col-0"}>Type：</div>
//                         <div className="flex-1 word-wrap">{item.type}</div>
//                     </div>
//                     <div className="row">
//                         <div className={style.exceptionLabel + " col-0"}>Time：</div>
//                         <div className="flex-1">{formatDate(new Date(item.timestamp / 1000))}</div>
//                     </div>
//                     <div className="row">
//                         <div className={style.exceptionLabel + " col-0"}>Stack：</div>
//                         <div className="flex-1">{item.stack?.length > 0 ?
//                             <HighlightCode rows={item.stack.split('\n')} title={spanName + '-' + item.message} /> : '无'}</div>
//                     </div>
//                     {/* {Object.keys(item).map((key: any) => (
//                         <ul key={key}>{key}：{item[key]}</ul>
//                     ))} */}
//                 </Collapse>
//             ))}
//         </>
//     );
// }
// const getStyles = (theme: GrafanaTheme2) => ({
//     exceptionLabel: css`
//     font-size: 14px;
//     font-weight: bold;
//     color: ${theme.colors.text.primary}
//     `
// })

// export default ExceptionsTable;
