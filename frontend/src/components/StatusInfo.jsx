import React from "react"
import { FaCircle } from "react-icons/fa";
import { StatusColorMap } from "src/constants";

function StatusInfo({ status }) {
  return <div className="p-2"><FaCircle color={StatusColorMap[status]}/></div>
}
export default StatusInfo
