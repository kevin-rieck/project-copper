export namespace engine {
	
	export class RegisterDefinition {
	    register: number;
	    count: number;
	    data_type: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.register = source["register"];
	        this.count = source["count"];
	        this.data_type = source["data_type"];
	    }
	}
	export class RegisterGroup {
	    id: string;
	    modbus_table: number;
	    definitions: RegisterDefinition[];
	
	    static createFrom(source: any = {}) {
	        return new RegisterGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.modbus_table = source["modbus_table"];
	        this.definitions = this.convertValues(source["definitions"], RegisterDefinition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Device {
	    id: string;
	    conn_id: string;
	    slave_id: number;
	    groups: Record<string, RegisterGroup>;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.conn_id = source["conn_id"];
	        this.slave_id = source["slave_id"];
	        this.groups = this.convertValues(source["groups"], RegisterGroup, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace main {
	
	export class ConfigLoadResult {
	    activeDeviceID: string;
	    deviceIDs: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigLoadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeDeviceID = source["activeDeviceID"];
	        this.deviceIDs = source["deviceIDs"];
	    }
	}

}

