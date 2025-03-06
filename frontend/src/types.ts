export interface BoilerData {
  id: number;
  reactorTemp: number;
  separatorTemp: number;
  furnaceTemp: number;
  condenserTemp: number;
  atmTemp: number;
  reactorPressure: number;
  gasTankPressure: number;
  // processStartTime: string;
  // timeOfReaction: string;
  // processEndTime: string;
  // coolingEndTime: string;
  nitrogenPurging: number;
  carbonDoorStatus: number;
  coCh4Leakage: number;
  jaaliBlockage: number;
  machineMaintenance: number;
  autoShutDown: number;
}
