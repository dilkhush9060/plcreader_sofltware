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
	export class Config {
	    plantId: string;
	    comPort: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plantId = source["plantId"];
	        this.comPort = source["comPort"];
	    }
	}

}

