/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

// import { useState, useEffect } from "react";
// import ReactFlow, {
//   ReactFlowProvider,
//   addEdge,
//   useNodesState,
//   useEdgesState,
//   MarkerType,
// } from "reactflow";
// import { graphlib, layout as dagreLayout } from "dagre";
// import * as d3 from "d3";
// import "reactflow/dist/style.css";
// import ArrowEdges from "components/ReactFlow/ArrowEdges";
// import ServiceNode from "components/ReactFlow/ServiceNode";
// // 节点和边的初始数据
// const initialNodes = [
//   {
//     id: "1",
//     data: { label: "Node 1", status: "error" },
//     position: { x: 0, y: 0 },
//     type: "serviceNode",
//   },
//   {
//     id: "2",
//     data: { label: "Node 2", status: "error" },
//     position: { x: 100, y: 0 },
//     type: "serviceNode",
//   },
//   {
//     id: "3",
//     data: { label: "Node 3", status: "error" },
//     position: { x: 0, y: 0 },
//     type: "serviceNode",
//   },
//   {
//     id: "4",
//     data: { label: "Node 4", status: "success" },
//     position: { x: 0, y: 0 },
//     type: "serviceNode",
//   },
// ];
// const nodeTypes = { serviceNode: ServiceNode };

// const markerEnd = {
//   type: MarkerType.ArrowClosed,
//   strokeWidth: 5,
//   width: 15,
//   height: 15,
// };
// const initialEdges = [
//   {
//     id: "e1-2",
//     source: "1",
//     target: "2",
//     markerEnd: markerEnd,
//   },
//   {
//     id: "e2-3",
//     source: "2",
//     target: "3",
//     markerEnd: markerEnd,
//   },
//   {
//     id: "e1-4",
//     source: "1",
//     target: "4",
//     markerEnd: markerEnd,
//   },
// ];
// const edgeTypes = { arrow: ArrowEdges };
// const dagreGraph = new graphlib.Graph();
// dagreGraph.setDefaultEdgeLabel(() => ({}));

// const nodeWidth = 172;
// const nodeHeight = 36;

// const getLayoutedElements = (nodes, edges) => {
//   dagreGraph.setGraph({ rankdir: "LR" }); // 自上而下的布局

//   nodes.forEach((node) => {
//     dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
//   });

//   edges.forEach((edge) => {
//     dagreGraph.setEdge(edge.source, edge.target);
//   });

//   dagreLayout(dagreGraph);

//   let xSum = 0,
//     ySum = 0;
//   nodes.forEach((node) => {
//     const nodeWithPosition = dagreGraph.node(node.id);
//     node.targetPosition = "left";
//     node.sourcePosition = "right";

//     // 更新节点位置
//     node.position = {
//       x: nodeWithPosition.x - nodeWidth / 2,
//       y: nodeWithPosition.y - nodeHeight / 2,
//     };

//     xSum += node.position.x;
//     ySum += node.position.y;
//   });

//   // 计算居中偏移
//   const xOffset = xSum / nodes.length;
//   const yOffset = ySum / nodes.length;

//   // 应用居中偏移
//   nodes.forEach((node) => {
//     node.position.x += xOffset;
//     node.position.y += yOffset;
//   });

//   return { nodes, edges };
// };

// const LayoutFlow = () => {
//   const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
//   const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

//   useEffect(() => {
//     const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
//       initialNodes,
//       initialEdges
//     );
//     setNodes([...layoutedNodes]);
//     setEdges([...layoutedEdges]);
//   }, []);

//   return (
//     <ReactFlowProvider>
//       <svg style={{ position: "absolute", top: 0, left: 0 }}>
//         <defs>
//           <marker
//             id="arrowhead"
//             viewBox="0 0 10 10"
//             refX="5"
//             refY="5"
//             markerWidth="6"
//             markerHeight="6"
//             orient="auto"
//           >
//             <path d="M0,0 L10,5 L0,10" fill="black" />
//           </marker>
//         </defs>
//       </svg>
//       <ReactFlow
//         nodes={nodes}
//         edges={edges}
//         onNodesChange={onNodesChange}
//         onEdgesChange={onEdgesChange}
//         nodeTypes={nodeTypes}
//       />
//     </ReactFlowProvider>
//   );
// };

// export default LayoutFlow;
