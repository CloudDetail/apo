export const serviceMock = [
    {
      name: "ts-travel-plan-service",
      status: "success",
      next: ["ts-route-plan-service", "ts-train-service", "ts-seat-service"],
    },
    {
      name: "ts-route-plan-service",
      status: "success",
      next: ["ts-travel-service", "ts-travel2-service"],
    },
    {
      name: "ts-travel-service",
      status: "success",
      next: ["ts-basic-service", "ts-seat-service", "ts-route-service"],
    },
    {
      name: "ts-travel2-service",
      status: "success",
      next: ["ts-basic-service", "ts-seat-service", "ts-route-service"],
    },
    {
      name: "ts-basic-service",
      status: "success",
      next: ["ts-station-service", "ts-train-service", "ts-route-service", "ts-price-service"],
    },
    {
      name: "ts-seat-service",
      status: "success",
      next: ["ts-order-service", "ts-config-service", "ts-order-other-service"],
    },
    {
      name: "ts-order-other-service",
      status: "success",
      next: [],
    },
  
    {
      name: "ts-route-service",
      status: "success",
      next: [],
    },
  
    {
      name: "ts-station-service",
      status: "success",
      next: [],
    },
    {
      name: "ts-train-service",
      status: "success",
      next: [],
    },
    {
      name: "ts-price-service",
      status: "success",
      next: [],
    },
    {
      name: "ts-order-service",
      status: "success",
      next: [],
    },
    {
      name: "ts-config-service",
      status: "success",
      next: [],
    },
    
  ]