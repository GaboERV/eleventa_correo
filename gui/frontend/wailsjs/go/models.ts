export namespace models {

	export class Config {
	    branch_name: string;
	    smtp_user: string;
	    smtp_pass: string;
	    target_email: string;
	    auto_day: string;
	    auto_time: string;
	    dll_dir: string;
	    origin_fdb: string;
	    temp_fdb: string;

	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch_name = source["branch_name"];
	        this.smtp_user = source["smtp_user"];
	        this.smtp_pass = source["smtp_pass"];
	        this.target_email = source["target_email"];
	        this.auto_day = source["auto_day"];
	        this.auto_time = source["auto_time"];
	        this.dll_dir = source["dll_dir"];
	        this.origin_fdb = source["origin_fdb"];
	        this.temp_fdb = source["temp_fdb"];
	    }
	}

}
