/* eslint-disable react/prop-types */
import React from "react";
import TableRow from "./tableRow";
import Empty from "../Empty/Empty";

function TableBody(props) {
  const { page, prepareRow, rowKey, loading,pageIndex ,pageSize,clickRow } = props.props;
  const getRowKeyValue = (row) => {
    if (!row) {
      return row.id;
    } else if (typeof rowKey === "function") {
      return rowKey(row.original);
    } else {
      return row.original[rowKey];
    }
  };
  return (
    <tbody>
      {(page &&
        page.length > 0 && 
        page.map((row, idx) => {
          prepareRow(row);
          return <TableRow row={row} key={getRowKeyValue(row)} clickRow={clickRow}/>;
        })) ||
        loading || <Empty />}
    </tbody>
  );
}

export default React.memo(TableBody);
