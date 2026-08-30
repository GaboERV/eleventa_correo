export namespace main {
	
	export class Branch {
	    branch_id: string;
	    nombre: string;
	    activa: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Branch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch_id = source["branch_id"];
	        this.nombre = source["nombre"];
	        this.activa = source["activa"];
	    }
	}

}

export namespace models {
	
	export class VentaDepartamento {
	    departamento_id: number;
	    departamento: string;
	    ventas_centavos: number;
	    ganancias_centavos: number;
	
	    static createFrom(source: any = {}) {
	        return new VentaDepartamento(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.departamento_id = source["departamento_id"];
	        this.departamento = source["departamento"];
	        this.ventas_centavos = source["ventas_centavos"];
	        this.ganancias_centavos = source["ganancias_centavos"];
	    }
	}
	export class Turno {
	    turno_id: number;
	    cajero_id: number;
	    cajero: string;
	    inicio_en: string;
	    termino_en: string;
	    sospechoso: boolean;
	    ventas_por_departamento: VentaDepartamento[];
	
	    static createFrom(source: any = {}) {
	        return new Turno(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.turno_id = source["turno_id"];
	        this.cajero_id = source["cajero_id"];
	        this.cajero = source["cajero"];
	        this.inicio_en = source["inicio_en"];
	        this.termino_en = source["termino_en"];
	        this.sospechoso = source["sospechoso"];
	        this.ventas_por_departamento = this.convertValues(source["ventas_por_departamento"], VentaDepartamento);
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
	export class Day {
	    fecha: string;
	    turnos: Turno[];
	
	    static createFrom(source: any = {}) {
	        return new Day(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fecha = source["fecha"];
	        this.turnos = this.convertValues(source["turnos"], Turno);
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
	export class Report {
	    schema_version: string;
	    branch_id: string;
	    hostname: string;
	    client_version: string;
	    generated_at: string;
	    dias: Day[];
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.branch_id = source["branch_id"];
	        this.hostname = source["hostname"];
	        this.client_version = source["client_version"];
	        this.generated_at = source["generated_at"];
	        this.dias = this.convertValues(source["dias"], Day);
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

