export namespace main {
	
	export class BoilerData {
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
	
	    static createFrom(source: any = {}) {
	        return new BoilerData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.reactorTemp = source["reactorTemp"];
	        this.separatorTemp = source["separatorTemp"];
	        this.furnaceTemp = source["furnaceTemp"];
	        this.condenserTemp = source["condenserTemp"];
	        this.atmTemp = source["atmTemp"];
	        this.reactorPressure = source["reactorPressure"];
	        this.gasTankPressure = source["gasTankPressure"];
	        this.processStartTime = source["processStartTime"];
	        this.timeOfReaction = source["timeOfReaction"];
	        this.processEndTime = source["processEndTime"];
	        this.coolingEndTime = source["coolingEndTime"];
	        this.nitrogenPurging = source["nitrogenPurging"];
	        this.carbonDoorStatus = source["carbonDoorStatus"];
	        this.coCh4Leakage = source["coCh4Leakage"];
	        this.jaaliBlockage = source["jaaliBlockage"];
	        this.machineMaintenance = source["machineMaintenance"];
	        this.autoShutDown = source["autoShutDown"];
	    }
	}

}

