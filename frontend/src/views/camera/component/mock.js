export const mockData = [
  {
    name: "ts-travel-service",
    status: "running",
    x: 300,
    y: 300,
    next: ["ts-basic-service", "ts-seat-service"],
  },
  {
    name: "ts-basic-service",
    status: "running",
    x: 800,
    y: 300,
    next: ["ts-station-service", "ts-train-service", "ts-route-service", "ts-price-service"],
  },
  {
    name: "ts-station-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },
  {
    name: "ts-train-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },
  {
    name: "ts-route-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },
  {
    name: "ts-price-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },

  {
    name: "ts-seat-service",
    status: "running",
    x: 550,
    y: 500,
    next: ["ts-order-service", "ts-config-service"],
  },

  {
    name: "ts-order-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },
  {
    name: "ts-config-service",
    status: "running",
    x: 550,
    y: 100,
    next: ["mysql"],
  },
];
