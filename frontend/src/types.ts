export interface BoilerData {
  id: number;
  reactorTemp: number;
  separatorTemp: number;
  furnaceTemp: number;
  condenserTemp: number;
  atmTemp: number;
  reactorPressure: number;
  gasTankPressure: number;
  processStartTime: number;
  timeOfReaction: number;
  processEndTime: number;
  coolingEndTime: number;
  nitrogenPurging: number;
  carbonDoorStatus: number;
  coCh4Leakage: number;
  jaaliBlockage: number;
  machineMaintenance: number;
  autoShutDown: number;
}
