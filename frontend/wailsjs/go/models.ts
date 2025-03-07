export namespace main {
	
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
	export class PLCDataResponse {
	    success: boolean;
	    data: number[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PLCDataResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.data = source["data"];
	        this.error = source["error"];
	    }
	}

}

