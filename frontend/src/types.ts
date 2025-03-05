export interface BoilerData {
  id: number;
  reactorTemp: number;
  separatorTemp: number;
  furnaceTemp: number;
  condenserTemp: number;
  atmTemp: number;
  reactorPressure: number;
  gasTankPressure: number;
  processStartTime: string;
  timeOfReaction: string;
  processEndTime: string;
  coolingEndTime: string;
  nitrogenPurging: string;
  carbonDoorStatus: string;
  coCh4Leakage: string;
  jaaliBlockage: string;
  machineMaintenance: string;
  autoShutDown: string;
}
