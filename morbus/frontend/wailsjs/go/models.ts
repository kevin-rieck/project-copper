export namespace engine {
	
	export class BulkCreateRegisterDefinitionsRequest {
	    device_id: string;
	    group_id: string;
	    start_register: number;
	    quantity: number;
	    data_type: string;
	    count: number;
	    byte_order: string;
	    name_pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new BulkCreateRegisterDefinitionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.start_register = source["start_register"];
	        this.quantity = source["quantity"];
	        this.data_type = source["data_type"];
	        this.count = source["count"];
	        this.byte_order = source["byte_order"];
	        this.name_pattern = source["name_pattern"];
	    }
	}
	export class BulkDeleteRegisterDefinitionsRequest {
	    device_id: string;
	    group_id: string;
	    definition_ids: string[];
	
	    static createFrom(source: any = {}) {
	        return new BulkDeleteRegisterDefinitionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.definition_ids = source["definition_ids"];
	    }
	}
	export class BulkEditRegisterDefinitionsRequest {
	    device_id: string;
	    group_id: string;
	    definition_ids: string[];
	    data_type: string;
	    count: number;
	    byte_order: string;
	
	    static createFrom(source: any = {}) {
	        return new BulkEditRegisterDefinitionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.definition_ids = source["definition_ids"];
	        this.data_type = source["data_type"];
	        this.count = source["count"];
	        this.byte_order = source["byte_order"];
	    }
	}
	export class CreateRegisterDefinitionRequest {
	    device_id: string;
	    group_id: string;
	    name: string;
	    register: number;
	    count: number;
	    data_type: string;
	    byte_order: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateRegisterDefinitionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.name = source["name"];
	        this.register = source["register"];
	        this.count = source["count"];
	        this.data_type = source["data_type"];
	        this.byte_order = source["byte_order"];
	    }
	}
	export class CreateRegisterGroupRequest {
	    device_id: string;
	    name: string;
	    modbus_table: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateRegisterGroupRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.name = source["name"];
	        this.modbus_table = source["modbus_table"];
	    }
	}
	export class DeleteRegisterDefinitionRequest {
	    device_id: string;
	    group_id: string;
	    definition_id: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteRegisterDefinitionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.definition_id = source["definition_id"];
	    }
	}
	export class DeleteRegisterGroupRequest {
	    device_id: string;
	    group_id: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteRegisterGroupRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	    }
	}
	export class RegisterDefinition {
	    id: string;
	    name: string;
	    register: number;
	    count: number;
	    data_type: string;
	    byte_order?: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.register = source["register"];
	        this.count = source["count"];
	        this.data_type = source["data_type"];
	        this.byte_order = source["byte_order"];
	    }
	}
	export class RegisterGroup {
	    id: string;
	    name: string;
	    modbus_table: number;
	    definitions: RegisterDefinition[];
	
	    static createFrom(source: any = {}) {
	        return new RegisterGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
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
	    byte_order: string;
	    groups: RegisterGroup[];
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.conn_id = source["conn_id"];
	        this.slave_id = source["slave_id"];
	        this.byte_order = source["byte_order"];
	        this.groups = this.convertValues(source["groups"], RegisterGroup);
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
	export class DuplicateRegisterDefinitionsRequest {
	    device_id: string;
	    source_group_id: string;
	    target_group_id: string;
	    definition_ids: string[];
	    address_offset: number;
	    name_pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new DuplicateRegisterDefinitionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.source_group_id = source["source_group_id"];
	        this.target_group_id = source["target_group_id"];
	        this.definition_ids = source["definition_ids"];
	        this.address_offset = source["address_offset"];
	        this.name_pattern = source["name_pattern"];
	    }
	}
	export class MoveRegisterDefinitionsRequest {
	    device_id: string;
	    source_group_id: string;
	    target_group_id: string;
	    definition_ids: string[];
	
	    static createFrom(source: any = {}) {
	        return new MoveRegisterDefinitionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.source_group_id = source["source_group_id"];
	        this.target_group_id = source["target_group_id"];
	        this.definition_ids = source["definition_ids"];
	    }
	}
	
	
	export class UpdateRegisterDefinitionRequest {
	    device_id: string;
	    group_id: string;
	    definition_id: string;
	    name: string;
	    register: number;
	    count: number;
	    data_type: string;
	    byte_order: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateRegisterDefinitionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.definition_id = source["definition_id"];
	        this.name = source["name"];
	        this.register = source["register"];
	        this.count = source["count"];
	        this.data_type = source["data_type"];
	        this.byte_order = source["byte_order"];
	    }
	}
	export class UpdateRegisterGroupRequest {
	    device_id: string;
	    group_id: string;
	    name: string;
	    modbus_table: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateRegisterGroupRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.group_id = source["group_id"];
	        this.name = source["name"];
	        this.modbus_table = source["modbus_table"];
	    }
	}

}

export namespace main {
	
	export class ConfigLoadResult {
	    loaded: boolean;
	    activeDeviceID: string;
	    deviceIDs: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigLoadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loaded = source["loaded"];
	        this.activeDeviceID = source["activeDeviceID"];
	        this.deviceIDs = source["deviceIDs"];
	    }
	}

}

