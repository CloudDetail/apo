/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// import {
//   CardHeader,
//   Card,
//   Grid,
//   RadioGroup,
//   FormControlLabel,
//   Radio,
//   ToggleButtonGroup,
//   ToggleButton,
//   Button,
//   Select,
//   MenuItem,
//   InputLabel,
//   FormControl,
// } from "@mui/material";
// import { useTheme } from "@emotion/react";
// import FlowChart from "./component/test";
// import Topology from "./component/Topology";
// import MDInput from "components/MDInput";
// import { useState } from "react";
// import { Height } from "@mui/icons-material";
// export default function Camera() {
//   const [controller] = useMaterialUIController();
//   const { darkMode } = controller;
//   console.log(darkMode);
//   const theme = useTheme();
//   console.log(theme); // 输出当前主题对象

//   const [topologyType, setTopologyType] = useState("all");
//   const [timeRange, setTimeRange] = useState("last 5 minutes");
//   const [age, setAge] = useState("last 5 minutes");
//   const handleAlignment = (event, newFormats) => {
//     setTopologyType(newFormats);
//   };
//   return (
//     <MDBox
//       sx={{ height: "calc(100vh - 100px)", bgColor: "red", width: "100%", overflow: "hidden" }}
//     >
//       <MDBox display="flex" sx={{ justifyContent: "space-between", my: 1 }}>
//         <MDBox display="flex" sx={{ flex: "auto", alignItems: "center" }}>
//           <MDInput label="Search here" sx={{ mx: 2 }} />
//           <ToggleButtonGroup
//             value={topologyType}
//             exclusive
//             color="primary"
//             onChange={handleAlignment}
//             aria-label="text alignment"
//           >
//             <ToggleButton value="all" aria-label="全链路拓扑图">
//               全链路拓扑图
//             </ToggleButton>
//             <ToggleButton value="simple" aria-label="精确展示">
//               精确展示
//             </ToggleButton>
//           </ToggleButtonGroup>
//         </MDBox>
//         <MDBox display="flex" sx={{ flex: "auto", justifyContent: "flex-end" }}>
//           {/* <MDInput value={timeRange} /> */}
//           <MDBox sx={{ minWidth: "120px", Height: "100%" }}>
//             <MDInput
//               size="large"
//               select
//               labelId="demo-simple-select-label"
//               id="demo-simple-select"
//               value={timeRange}
//               label="时间"
//               fullWidth
//             >
//               {/* <MenuItem value="Male">Male</MenuItem>
//               <MenuItem value="Female">Female</MenuItem> */}
//             </MDInput>
//           </MDBox>
//         </MDBox>
//       </MDBox>
//       <Grid container className="h-full">
//         <Grid item xs={topologyType === "all" ? 12 : 6}>
//           <MDBox bgColor={darkMode ? "transparent" : "grey-100"} className="w-full h-full">
//             <Card variant="outlined" sx={{ height: "100%", width: "100%" }}>
//               {/* <CardHeader title="业务拓扑图"></CardHeader> */}
//               {/* <FlowChart /> */}
//               <Topology />
//             </Card>
//           </MDBox>
//         </Grid>
//         <Grid item xs={topologyType === "all" ? 0 : 6}>
//           <MDBox bgColor={darkMode ? "transparent" : "grey-100"} className="w-full h-full">
//             <Card variant="outlined" sx={{ height: "100%", width: "100%" }}>
//               {/* <CardHeader title="业务拓扑图"></CardHeader> */}
//               {/* <FlowChart /> */}
//               <Topology />
//             </Card>
//           </MDBox>
//         </Grid>
//       </Grid>
//     </MDBox>
//   );
// }
