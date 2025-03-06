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

}

