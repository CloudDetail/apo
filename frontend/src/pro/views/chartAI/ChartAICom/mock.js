import topologyPng from 'src/pro/assets/topology.png'
import eventListPng from 'src/pro/assets/eventList.png'
import timelinePng from 'src/pro/assets/timeline.png'
import newEventPng from 'src/pro/assets/newEvent.png'
import resolveEventPng from 'src/pro/assets/resolveEvent.png'
import polariPng from 'src/pro/assets/polari.png'
import mock1Png from 'src/pro/assets/1.png'
import mock2Png from 'src/pro/assets/2.png'
import mock5Png from 'src/pro/assets/5.png'
import mock61Png from 'src/pro/assets/6-1.png'
import mock62Png from 'src/pro/assets/6-2.png'
import mock7Png from 'src/pro/assets/7.png'
import mock9_1Png from 'src/pro/assets/9-1.png'
import mock9_2Png from 'src/pro/assets/9-2.png'
import mock10Png from 'src/pro/assets/10.png'

import { Image, Table } from 'antd'
function createDownloadLink(data, fileName) {
  // 检查 data 是否是字符串或者对象
  const content =
    typeof data === 'string'
      ? data // 如果是字符串，直接使用
      : JSON.stringify(data, null, 2) // 如果是对象，序列化为 JSON 格式

  // 创建 Blob 并生成下载链接
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)

  // 返回 HTML 链接
  return `<a href="${url}" download="${fileName}">${fileName}</a>`
}
const topologyMock = `ts-gateway-service|POST travel-plan
  └──ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest
    └──ts-train-service|GET /api/v1/trainservice/trains/byName/{name}
    └──ts-seat-service|POST /api/v1/seatservice/seats/left_tickets
      └──ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets
      └──ts-order-service|POST /api/v1/orderservice/order/tickets
      └──ts-config-service|GET /api/v1/configservice/configs/{configName}
    └──ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute
      └──ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}
        └──ts-route-service|GET /api/v1/routeservice/routes/{routeId}
      └──ts-travel-service|GET /api/v1/travelservice/routes/{tripId}
        └──ts-route-service|GET /api/v1/routeservice/routes/{routeId}
      └──ts-travel-service|POST /api/v1/travelservice/trips/left
        └──ts-basic-service|POST /api/v1/basicservice/basic/travels
          └──ts-train-service|POST /api/v1/trainservice/trains/byNames
          └──ts-station-service|POST /api/v1/stationservice/stations/idlist
          └──ts-price-service|POST /api/v1/priceservice/prices/byRouteIdsAndTrainTypes
          └──ts-route-service|POST /api/v1/routeservice/routes/byIds/
        └──ts-seat-service|POST /api/v1/seatservice/seats/left_tickets
          └──ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets
          └──ts-order-service|POST /api/v1/orderservice/order/tickets
          └──ts-config-service|GET /api/v1/configservice/configs/{configName}
      └──ts-travel2-service|POST /api/v1/travel2service/trips/left
        └──ts-basic-service|POST /api/v1/basicservice/basic/travels
          └──ts-train-service|POST /api/v1/trainservice/trains/byNames
          └──ts-station-service|POST /api/v1/stationservice/stations/idlist
          └──ts-price-service|POST /api/v1/priceservice/prices/byRouteIdsAndTrainTypes
          └──ts-route-service|POST /api/v1/routeservice/routes/byIds/
        └──ts-seat-service|POST /api/v1/seatservice/seats/left_tickets
          └──ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets
          └──ts-order-service|POST /api/v1/orderservice/order/tickets
          └──ts-config-service|GET /api/v1/configservice/configs/{configName}`
const eventListMock = {
  'ts-gateway-service|POST travel-plan': {
    5: {
      add: 21,
      duplicate: 0,
      resolve: 21,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
    1: {
      add: 4,
      duplicate: 12,
      resolve: 4,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774335000,
    },
  },
  'ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest': {
    5: {
      add: 16,
      duplicate: 0,
      resolve: 16,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
    1: {
      add: 5,
      duplicate: 14,
      resolve: 5,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774335000,
    },
  },
  'ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute': {
    5: {
      add: 14,
      duplicate: 0,
      resolve: 14,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
    1: {
      add: 3,
      duplicate: 15,
      resolve: 3,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774335000,
    },
  },
  'ts-travel2-service|POST /api/v1/travel2service/trips/left': {
    5: {
      add: 14,
      duplicate: 0,
      resolve: 14,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
    2: {
      add: 1,
      duplicate: 0,
      resolve: 1,
      keep: 0,
      lastTime: 1731555300824000,
      firstTime: 1731555250820000,
    },
    1: {
      add: 2,
      duplicate: 14,
      resolve: 2,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774335000,
    },
  },
  'ts-seat-service|POST /api/v1/seatservice/seats/left_tickets': {
    5: {
      add: 14,
      duplicate: 0,
      resolve: 14,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
    1: {
      add: 3,
      duplicate: 9,
      resolve: 4,
      keep: 0,
      lastTime: 1731556214498000,
      firstTime: 1731554774335000,
    },
  },
  'ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets': {
    5: {
      add: 14,
      duplicate: 0,
      resolve: 14,
      keep: 0,
      lastTime: 1731555953164000,
      firstTime: 1731555719310000,
    },
  },
  'ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}': {
    2: {
      add: 1,
      duplicate: 0,
      resolve: 1,
      keep: 0,
      lastTime: 1731555300824000,
      firstTime: 1731555250820000,
    },
    1: {
      add: 3,
      duplicate: 10,
      resolve: 3,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774334000,
    },
  },
  'ts-route-service|GET /api/v1/routeservice/routes/{routeId}': {
    1: {
      add: 1,
      duplicate: 7,
      resolve: 2,
      keep: 0,
      lastTime: 1731556154487000,
      firstTime: 1731554774334000,
    },
  },
  'ts-travel-service|GET /api/v1/travelservice/routes/{tripId}': {
    1: {
      add: 1,
      duplicate: 7,
      resolve: 2,
      keep: 0,
      lastTime: 1731556154488000,
      firstTime: 1731554774334000,
    },
  },
  'ts-basic-service|POST /api/v1/basicservice/basic/travels': {
    1: {
      add: 1,
      duplicate: 11,
      resolve: 2,
      keep: 0,
      lastTime: 1731556394526000,
      firstTime: 1731554774335000,
    },
  },
  'ts-route-service|POST /api/v1/routeservice/routes/byIds/': {
    1: {
      add: 0,
      duplicate: 10,
      resolve: 1,
      keep: 0,
      lastTime: 1731556154488000,
      firstTime: 1731554774335000,
    },
  },
  'ts-travel-service|POST /api/v1/travelservice/trips/left': {
    1: {
      add: 1,
      duplicate: 13,
      resolve: 1,
      keep: 1,
      lastTime: 1731556454533000,
      firstTime: 1731554774335000,
    },
  },
}
const delayMock = [
  {
    cpu_time: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        25102142.85714286, 26335384.615384616, 24088000.0, 23794000.0, 24259500.0,
        25197857.14285714, 24149250.0, 25443125.0, 24201473.684210528, 25006428.571428575,
        26006111.111111116, 25526875.0, 25054285.714285716, 26415555.55555556, 24591666.666666664,
        26155000.0, 24150720.0, 25102142.85714286, 21373058.82352941, 21598800.000000004,
        21153000.000000004, 24276461.53846154, 28524833.333333336, 27104166.666666668, 25987500.0,
        24186000.0, 24748000.000000007, 25359375.0, 25708333.333333336, 23985545.454545453,
        25217000.000000004,
      ],
    },
    disk_io_time: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        9.000000000000001e-10, 9.157894736842107e-10, 9.200000000000001e-10, 9.461538461538464e-10,
        9.000000000000002e-10, 9.000000000000001e-10, 9.490909090909093e-10, 9.500000000000002e-10,
        9.000000000000001e-10, 9.000000000000001e-10, 9.155172413793105e-10, 9.339622641509436e-10,
        9.321428571428571e-10, 9.157894736842107e-10, 9.000000000000001e-10, 9.152542372881357e-10,
        9.157894736842107e-10, 9.000000000000001e-10, 9.46551724137931e-10, 4705.300000000006,
        9.48214285714286e-10, 9.155172413793105e-10, 9.000000000000001e-10, 9.315789473684213e-10,
        9.346153846153848e-10, 9.169811320754719e-10, 9.333333333333334e-10, 9.155172413793105e-10,
        9.500000000000002e-10, 9.32727272727273e-10, 9.000000000000001e-10,
      ],
    },
    lock_gc_time: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        5491000.000000005, 6158600.000000008, 5779000.000000005, 5243350.000000001,
        7848600.000000009, 4958500.000000002, 2820899.9999999995, 3806333.333333335,
        3438480.000000001, 3773666.6666666684, 4102600.000000002, 3937000.0000000014,
        6049533.333333337, 4418800.000000004, 6191320.000000001, 4363000.000000001,
        4705300.000000004, 5491000.000000009, 3356400.0, 885905.0000000006, 2481500.000000005,
        3618500.0000000023, 4111900.0000000005, 5132575.000000002, 6158600.000000002,
        4418800.000000003, 4895200.000000005, 4705300.000000006, 3912500.0000000023, 6117700.0,
        6294933.333333335,
      ],
    },
    network_time: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        761280000.0000001, 784860000.0000002, 1245200000.0000005, 1288880000.0, 1356533333.333334,
        1013600000.0000001, 767536363.6363636, 690600000.0000002, 718500000.0, 715400000.0000001,
        665758000.0000001, 727800000.0000002, 722220000.0000002, 718500000.0000002,
        671951428.5714287, 632220000.0, 648580000.0000001, 721600000.0000005, 614223999.9999999,
        552700000.0000001, 529900000.00000024, 673938000.0, 650216000.0, 685950000.0000005,
        903920000.0000005, 842940000.0000004, 769227272.7272727, 685950000.0000005,
        820060000.0000004, 784860000.0000005, 706875000.0000001,
      ],
    },
    scheduling_time: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        6240400.00000001, 4768600.000000011, 3464400.0000000014, 3143050.000000002,
        4139800.0000000047, 3486000.000000002, 3692000.000000002, 2640625.000000001, 2408800.0,
        3076725.0, 2820900.0000000033, 2972500.0000000065, 2515000.0000000056, 3888000.000000002,
        2330400.0000000014, 3162000.0, 2242200.0000000005, 2947233.3333333344, 4035000.0000000023,
        2820900.0000000056, 4139800.000000001, 3004985.7142857146, 4139800.0000000047,
        4663100.000000002, 5190600.000000002, 3790000.0000000014, 4344400.000000001,
        3618500.0000000023, 2947233.3333333354, 4736950.000000002, 3399600.0000000014,
      ],
    },
    total_latency: {
      time: [
        'Thu, 14 Nov 2024 03:24:49 GMT',
        'Thu, 14 Nov 2024 03:25:49 GMT',
        'Thu, 14 Nov 2024 03:26:49 GMT',
        'Thu, 14 Nov 2024 03:27:49 GMT',
        'Thu, 14 Nov 2024 03:28:49 GMT',
        'Thu, 14 Nov 2024 03:29:49 GMT',
        'Thu, 14 Nov 2024 03:30:49 GMT',
        'Thu, 14 Nov 2024 03:31:49 GMT',
        'Thu, 14 Nov 2024 03:32:49 GMT',
        'Thu, 14 Nov 2024 03:33:49 GMT',
        'Thu, 14 Nov 2024 03:34:49 GMT',
        'Thu, 14 Nov 2024 03:35:49 GMT',
        'Thu, 14 Nov 2024 03:36:49 GMT',
        'Thu, 14 Nov 2024 03:37:49 GMT',
        'Thu, 14 Nov 2024 03:38:49 GMT',
        'Thu, 14 Nov 2024 03:39:49 GMT',
        'Thu, 14 Nov 2024 03:40:49 GMT',
        'Thu, 14 Nov 2024 03:41:49 GMT',
        'Thu, 14 Nov 2024 03:42:49 GMT',
        'Thu, 14 Nov 2024 03:43:49 GMT',
        'Thu, 14 Nov 2024 03:44:49 GMT',
        'Thu, 14 Nov 2024 03:45:49 GMT',
        'Thu, 14 Nov 2024 03:46:49 GMT',
        'Thu, 14 Nov 2024 03:47:49 GMT',
        'Thu, 14 Nov 2024 03:48:49 GMT',
        'Thu, 14 Nov 2024 03:49:49 GMT',
        'Thu, 14 Nov 2024 03:50:49 GMT',
        'Thu, 14 Nov 2024 03:51:49 GMT',
        'Thu, 14 Nov 2024 03:52:49 GMT',
        'Thu, 14 Nov 2024 03:53:49 GMT',
        'Thu, 14 Nov 2024 03:54:49 GMT',
      ],
      value: [
        784860000.0000002, 832380000.0000001, 1268600000.0000002, 1289771428.5714288,
        1356533333.333334, 1017000000.0, 768576923.0769231, 741750000.0, 734775000.0000001,
        732450000.0000002, 673120000.0, 791900000.0000001, 727800000.0, 746400000.0000002,
        737100000.0, 669030000.0000001, 675350909.0909091, 795420000.0000001, 652670000.0,
        570700000.0000001, 601545000.0, 723814285.7142859, 673120000.0, 690599999.9999999,
        943953333.3333337, 850332000.0000004, 781340000.0000004, 755700000.0, 835020000.0000004,
        848220000.0000007, 765000000.0,
      ],
    },
  },
]
export const mockDialogue = [
  {
    conversation: [
      {
        role: 'human',
        message: `拓扑图数据如文件 ${createDownloadLink(topologyMock, '拓扑图数据.json')}`,
        files: [
          {
            name: '拓扑图数据',
          },
        ],
      },
      {
        role: 'ai',
        message: '已记住微服务拓扑结构。',
      },
    ],
    explanation: (
      <div>
        <div>拓扑示例如下：通过告警深入分析可以看到告警关联的完整业务调用关系。</div>
        <Image src={mock1Png} width={'60%'} />
      </div>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message: `告警事件数据如下 ${createDownloadLink(eventListMock, '告警事件数据.json')}`,
      },
      {
        role: 'ai',
        message: `已记住告警事件数据。\n如果需要分析或报告这些数据，请提供进一步的指导。`,
      },
    ],
    explanation: (
      <>
        <div>
          示例：在告警深入分析中，可以通过拖动时间范围可以看到不同时间段的告警详细信息。智能对话中，提交的是这段时间所有的告警信息，并且每个告警进行已经被关联至应用接口
        </div>
        <Image src={mock2Png} width={'90%'} />
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message: `请根据上文提供的微服务拓扑图和每个节点的告警事件, 按照下文提供的规则找出符合规则的根因节点。
根因节点推导流程
1. 根据告警事件中异常类型和发生次数，记录异常事件从下游往上游传递的顺序
2. 每次给出一个规则后，回答满足该规则的节点，回答过程不用附带完整数据，总结数据即可
注意：在拓扑中，一个上游可能有多个下游节点，多个下游视为同级节点。如果同级节点中的服务都满足异常向上传递，那么这些下游节点都可认为是根因，且排除上游节点，不要遗漏节点。同时记得排除入口节点。`,
      },
      {
        role: 'ai',
        message: `了解根因节点推导流程。请提供规则，我将根据微服务拓扑图和告警事件数据，找出满足该规则的根因节点。在您提供规则之前，以下是推导过程中的关键点：

- 异常事件从下游往上游传递。
- 同级节点中，如果所有服务都满足异常传递条件，这些节点都可能是根因。
- 排除入口节点。
- 排除已经确定不是根因的上游节点。

请提供具体规则，以便开始分析。`,
      },
    ],
    explanation: <>先请LLM根据拓扑结构和告警信息完成疑似根因节点的识别。</>,
  },
  {
    conversation: [
      {
        role: 'human',
        message: `规则1:告警异常类型中值含义 1为应用异常,2为容器异常,3为基础设施异常,4为网络程序异常,5为JAVA堆栈抛出异常。没有告警事件的节点可以排除。`,
      },
      {
        role: 'ai',
        message:
          `
根据规则1，我们将关注告警异常类型为1（应用异常）、2（容器异常）、3（基础设施异常）、4（网络程序异常）、5（JAVA堆栈抛出异常）的节点。没有告警事件的节点将被排除。\n
根据提供的告警事件数据，以下是存在上述类型告警的节点：\n` +
          '- `ts-gateway-service|POST travel-plan`：存在类型1的告警。\n- `ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest`：存在类型1的告警\n\n' +
          '- `ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`：存在类型1的告警。\n' +
          '- `ts-travel2-service|POST /api/v1/travel2service/trips/left`：存在类型1和类型2的告警。\n' +
          '- `ts-seat-service|POST /api/v1/seatservice/seats/left_tickets`：存在类型1的告警。\n' +
          '- `ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets`：存在类型5的告警。\n' +
          '- `ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}`：存在类型1和类型2的告警。\n' +
          '- `ts-route-service|GET /api/v1/routeservice/routes/{routeId}`：存在类型1的告警。\n' +
          '- `ts-travel-service|GET /api/v1/travelservice/routes/{tripId}`：存在类型1的告警。\n' +
          '- `ts-basic-service|POST /api/v1/basicservice/basic/travels`：存在类型1的告警。\n' +
          '- `ts-route-service|POST /api/v1/routeservice/routes/byIds/`：存在类型1的告警。\n' +
          '- `ts-travel-service|POST /api/v1/travelservice/trips/left`：存在类型1的告警。\n' +
          `\n` +
          '\n由于入口节点`ts-gateway-service|POST travel-plan`存在告警，根据规则我们排除该节点。现在，我们需要检查每个节点的下游节点，以确定哪些节点可能是根因。\n' +
          '- `ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest`的下游节点是`ts-train-service`和`ts-seat-service`，但这两个节点在提供的告警事件数据中没有被提及，因此我们可以排除它们。\n' +
          '- `ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`的下游节点是`ts-travel2-service`和`ts-travel-service`，这两个节点都存在类型1的告警。\n' +
          '- `ts-travel2-service|POST /api/v1/travel2service/trips/left`和`ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}`都存在类型1的告警。\n' +
          '- `ts-seat-service|POST /api/v1/seatservice/seats/left_tickets`存在类型1的告警。\n' +
          '- `ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets`仅存在类型5的告警，不满足规则1中的其他类型。\n' +
          '- `ts-route-service|GET /api/v1/routeservice/routes/{routeId}`、`ts-travel-service|GET /api/v1/travelservice/routes/{tripId}`、`ts-basic-service|POST /api/v1/basicservice/basic/travels`和`ts-route-service|POST /api/v1/routeservice/routes/byIds/`都存在类型1的告警。\n' +
          '- `ts-travel-service|POST /api/v1/travelservice/trips/left`存在类型1的告警。\n' +
          '综上所述，可能存在根因的节点包括：\n' +
          '- `ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`\n' +
          '- `ts-travel2-service|POST /api/v1/travel2service/trips/left`\n' +
          '- `ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}`\n' +
          '- `ts-seat-service|POST /api/v1/seatservice/seats/left_tickets`\n' +
          '- `ts-route-service|GET /api/v1/routeservice/routes/{routeId}`\n' +
          '- `ts-travel-service|GET /api/v1/travelservice/routes/{tripId}`\n' +
          '- `ts-basic-service|POST /api/v1/basicservice/basic/travels`\n' +
          '- `ts-route-service|POST /api/v1/routeservice/routes/byIds/`\n' +
          '- `ts-travel-service|POST /api/v1/travelservice/trips/left`\n' +
          '请注意，这些节点是基于提供的告警事件数据进行分析的，并且我们假设只有当所有同级节点的服务都满足异常传递条件时，它们才可能是根因。如果需要进一步的规则或条件，请提供。',
      },
    ],
    explanation: (
      <>
        告警信息详细解释：
        <ol style={{ listStyleType: 'disc', paddingLeft: 20 }}>
          <li>应用层告警（URL接口延时异常和错误异常）</li>
          <li>网络层告警（节点之间连接的网络周期性Ping值异常）</li>
          <li>基础设施主机层告警（主机的CPU、 内存、磁盘、网卡吞吐量异常）</li>
          <li>容器告警（k8s事件比如容器killed， 容器重启等异常）</li>
          <li>
            错误告警（Java程序使用java
            Exception，非Java程序，使用日志错误数异常。解决微服务场景中会识别不了错误的场景、将错误封装到error
            code中，导致http status code永远是200）
          </li>
        </ol>
        专家经验：只有应用层告警和错误告警能够在调用链路上传递。
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message: `规则2:在拓扑中关系中，当下游节点发生异常告警，上游节点之后发生应用的异常告警或错误异常，那么认为该类型的告警从下游向上游传递，还可以向更上游传递。因此根因节点需要满足发生异常告警，且需要查看其上游节点是否满足异常向上传递的条件。上游节点的应用告警是由下游传递导致的，所以应该忽略上游节点，但是上游节点有非应用告警的除外。`,
      },
      {
        role: 'ai',
        message:
          '根据提供的告警事件数据，我们可以检查以下关系：\n' +
          '- `ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest`的下游节点是`ts-train-service`和`ts-seat-service`，但这两个节点在提供的告警事件数据中没有被提及，因此我们可以排除它们。\n' +
          '- `ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`的下游节点是`ts-travel2-service`和`ts-travel-service`，这两个节点都存在类型1的告警，而`ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`本身也存在类型1的告警。因此，异常从下游传递到了`ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`。\n' +
          '- `ts-travel2-service|POST /api/v1/travel2service/trips/left`的下游节点是`ts-basic-service`，但`ts-basic-service`在提供的告警事件数据中没有类型1的告警，因此我们不能确定异常是否从`ts-travel2-service|POST /api/v1/travel2service/trips/left`传递到了`ts-basic-service`。\n' +
          '- `ts-seat-service|POST /api/v1/seatservice/seats/left_tickets`存在类型1的告警，但没有提到其下游节点的告警情况，因此我们无法确定异常是否传递到了更下游的节点。\n' +
          '- `ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets`存在类型5的告警，但这是JAVA堆栈抛出异常，不满足规则2的应用异常告警条件。\n' +
          '- 其他节点的下游节点在提供的告警事件数据中没有被提及，因此我们无法确定异常是否从这些节点传递到了下游节点。\n' +
          '\n根据规则2，我们可以得出以下结论：\n' +
          '- `ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute`可能是根因节点，因为它的下游节点存在类型1的告警，并且它自己也存在类型1的告警，表明异常从下游传递到了这个节点。\n请提供进一步的规则或条件，以便进行更深入的分析。',
      },
    ],
    explanation: (
      <>
        <div>
          <div className="w-full">
            <div>如下图示例，应用层告警和错误告警都是从下游往上游传递。</div>
            <div>上游节点的应用告警是由下游传递导致的，所以上游节点不是疑似根因节点</div>
            <div>
              {`示例中应用告警传递过程 ts-route-service -> ts-travel-service -> ts-route-plan-service
              -> ts-gateway-servcie。`}
            </div>
          </div>
          <Image src={mock5Png} width={'100%'} />
        </div>
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message: `规则3:根因节点需要满足异常从该节点开始向上游传递递，同时该类型异常告警事件是持续发生的，而不是偶发的（可以根据告警是否存留,告警事件次数和告警发生时间判断）。如果节点的异常告警是偶发，那么它不应该被认定为根因节点
根据规则3，我们需要确定根因节点是否满足以下条件：
- 异常从该节点开始向上游传递。
- 该类型异常告警事件是持续发生的，而不是偶发的。`,
      },
      {
        role: 'ai',
        message: `我们可以通过以下方式来判断异常是否持续发生：
- 检查告警事件是否存留（\`keep\`字段）。
- 检查告警事件的次数（\`add\`字段）。
- 检查告警事件的发生时间（\`firstTime\`和\`lastTime\`字段），以确定是否有连续发生的情况。

根据提供的告警事件数据，我们可以分析以下节点：

- \`ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute\`：存在类型1的告警，但是所有类型1的告警都已经被解决（\`resolve\`字段），这意味着这些告警是偶发的，而不是持续发生的。

其他节点在提供的告警事件数据中没有提到具体的告警存留情况或发生次数，因此我们无法直接判断它们是否满足规则3。

然而，我们可以根据规则3的指导原则，排除那些没有持续发生告警的节点。由于\`ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute\`的类型1告警已经被解决，它不满足规则3的条件，因此不能被认定为根因节点。

为了找到满足规则3的根因节点，我们需要进一步的信息来确定其他节点的告警是否持续发生。如果其他节点的告警是持续发生的，并且满足规则1和规则2，那么它们可能是根因节点。
目前，我们没有足够的信息来确定具体的根因节点。如果有更多的告警事件数据或其他规则，请提供，以便我们能够继续分析。`,
      },
    ],
    // explanation: (
    //   <div>
    //     <div>偶发的是噪音，应该排除。例如在这段时间内，在告警新增之后，很快的看到了已经解决。</div>
    //     <div>新增发生在：11：19：35秒 </div>
    //     <Image src={newEventPng} width={'70%'} />
    //     <div>已解决发生在：11：20：35秒</div>
    //     <Image src={resolveEventPng} width={'70%'} />
    //   </div>
    // ),
    explanation: (
      <>
        <div>如果应用接口产生的告警是偶发的，应该排除出疑似根因节点以外。</div>
        <div>偶发的告警定义：在故障排查时间段内，告警新增之后，很快的看到了已经解决。</div>
        <div>新增发生在：11：19：35秒</div>
        <Image src={mock61Png} width={'100%'} />

        <div>已解决发生在：11：20：35秒</div>
        <Image src={mock62Png} width={'100%'} />
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message:
          '规则4:如果某个节点发生异常，但是上游节点未发生应用异常告警事件，可以将这个节点及其分支全部排除',
      },
      {
        role: 'ai',
        message: `结合之前的分析，我们可以应用这个规则来排除一些节点：

- \`ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets\`：虽然这个节点发生了类型5的告警（JAVA堆栈抛出异常），但是它的上游节点\`ts-seat-service|POST /api/v1/seatservice/seats/left_tickets\`在提供的告警事件数据中没有发生类型1的应用异常告警。因此，根据规则4，我们可以排除\`ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets\`及其分支。

对于其他节点，我们需要确认它们的上游节点是否发生了应用异常告警：

- \`ts-travel2-service|POST /api/v1/travel2service/trips/left\`和\`ts-travel2-service|GET /api/v1/travel2service/routes/{tripId}\`：它们的上游节点是\`ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute\`，这个节点发生了类型1的应用异常告警。因此，这些节点暂时不被排除。
- \`ts-seat-service|POST /api/v1/seatservice/seats/left_tickets\`：它的上游节点是\`ts-travel-plan-service|POST /api/v1/travelplanservice/travelPlan/cheapest\`，这个节点也发生了类型1的应用异常告警。因此，\`ts-seat-service|POST /api/v1/seatservice/seats/left_tickets\`暂时不被排除。
- \`ts-route-service|GET /api/v1/routeservice/routes/{routeId}\`、\`ts-travel-service|GET /api/v1/travelservice/routes/{tripId}\`、\`ts-basic-service|POST /api/v1/basicservice/basic/travels\`和\`ts-route-service|POST /api/v1/routeservice/routes/byIds/\`：这些节点的上游节点在提供的告警事件数据中没有被提及，因此我们无法确定它们是否发生了应用异常告警。

根据规则4，我们可以排除\`ts-order-other-service|POST /api/v1/orderOtherService/orderOther/tickets\`及其分支。其他节点的排除情况需要进一步的信息来确定。如果提供了更多的告警事件数据或其他规则，请提供，以便我们能够继续分析。
`,
      },
    ],
    // explanation: (
    //   <div>
    //     <div>偶发的是噪音，应该排除。例如在这段时间内，在告警新增之后，很快的看到了已经解决。</div>
    //     <div>新增发生在：11：19：35秒 </div>
    //     <Image src={newEventPng} width={'70%'} />
    //     <div>已解决发生在：11：20：35秒</div>
    //     <Image src={resolveEventPng} width={'70%'} />
    //   </div>
    // ),

    explanation: (
      <>
        <div>根因节点产生的告警，应该沿着调用链路逆向传递到业务入口。</div>
        <div>示例中，告警只传递到了某个上游就断掉了，说明该上游和下游告警节点都不是根因节点。</div>
        <Image src={mock7Png} width={'100%'} />
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message:
          '汇总之前的节点信息，返回一个最可能是根因的节点,回答格式为造成根因的是xxx节点,不需要其他任何额外信息',
      },
      {
        role: 'ai',
        message: `造成根因的是ts-route-plan-service|POST /api/v1/routeplanservice/routePlan/cheapestRoute节点。`,
      },
    ],
    explanation: (
      <>
        <div>汇总</div>
        <div>
          不同模型的推理准确度并不相同，请根据您使用的模型判断需不需要人为使用“告警深入分析功能”确认疑似根因节点。
        </div>
        <div>
          目前测试数据如下：
          <Table
            pagination={false}
            dataSource={[
              {
                key: '1',
                name: 'GPT4o mini',
                value: '90%',
              },
              {
                key: '2',
                name: '智普GLM',
                value: '85%',
              },
              {
                key: '2',
                name: '通义千问',
                value: '80%',
              },
            ]}
            columns={[
              {
                title: '模型',
                dataIndex: 'name',
                key: 'name',
              },
              {
                title: '准确率',
                dataIndex: 'value',
                key: 'value',
              },
            ]}
          />
        </div>
      </>
    ),
  },
  {
    conversation: [
      {
        role: 'human',
        message: `该节点的延时曲线的数据
第一层KEY为延时类型,其中total_latency为总延时, lock_gc_time, network_time, scheduling_time, cpu_time, disk_io_time 为不同类型的曲线
第二层KEY中time为曲线时间, vaule对应每个时间点的延时数据
请根据曲线变化的相似性找出和 total_latency 变化最相似的曲线 (不要根据值的大小), 然后给出类型名称即可。
回答格式为延时曲线中的xx类型最相似,xx内容且必须为(lock_gc_time, network_time, scheduling_time, cpu_time, disk_io_time类型中的其中一个)
延时曲线数据如下 ${createDownloadLink(delayMock, '延时曲线数据.json')}
`,
      },
      {
        role: 'ai',
        message: `延时曲线中的scheduling_time类型最相似。
        <div style="overflow-x: auto;">
        
| 本次请求耗时                                                 | 请求时间            | 入口应用名称       | 入口服务端点     | 查看报告     |
| ------------------------------------------------------------ | ------------------- | ------------------ | ---------------- | ------------ |
| <div>1219ms<span class="ant-tag ant-tag-red css-1leukm8" style="margin-left: 5px;">偏离基线0.11%</span></div> | 2024-11-15 16:41:13 | ts-gateway-service | POST travel-plan | [查看](http://kindling.myddns.me:31553/#/cause/report/13b76c4d8c129bb6b241d552cbc80048/9465c9c9999bceee?mutatedType=scheduling_time) |
| <div>1219ms<span class="ant-tag ant-tag-red css-1leukm8" style="margin-left: 5px;">偏离基线0.21%</span></div> | 2024-11-15 16:42:25 | ts-gateway-service | POST travel-plan | [查看](http://kindling.myddns.me:31553/#/cause/report/e3675e18900c13d6209e628ca67799ab/25877bdd0bada47f?mutatedType=scheduling_time) |
</div>`,
      },
    ],
    explanation: (
      <>
        <div>
          北极星指标曲线是利用eBPF技术，将请求延时完整的拆解成了主要的几个故障方向，当延时变化之时，可以根据哪个分项的变化曲线最相似，从而判断疑似故障节点的故障方向。
        </div>
        <Image src={mock9_1Png} />
        <Image src={mock9_2Png} />
        <div>最后可以查看故障方向的进一步分析报告。</div>
        <Image src={mock10Png} />
      </>
    ),
  },
]
