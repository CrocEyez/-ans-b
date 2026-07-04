export namespace main {
	
	export class Account {
	    id: number;
	    username: string;
	    nickname: string;
	    role: string;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	        this.role = source["role"];
	    }
	}
	export class HotQuestion {
	    id: number;
	    question: string;
	    access_count: number;
	
	    static createFrom(source: any = {}) {
	        return new HotQuestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.question = source["question"];
	        this.access_count = source["access_count"];
	    }
	}
	export class HotQuestionsStatus {
	    available: boolean;
	    message: string;
	    questions?: HotQuestion[];
	
	    static createFrom(source: any = {}) {
	        return new HotQuestionsStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.message = source["message"];
	        this.questions = this.convertValues(source["questions"], HotQuestion);
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
	export class LoginResult {
	    token: string;
	    expires_in: number;
	    user: Account;
	
	    static createFrom(source: any = {}) {
	        return new LoginResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.expires_in = source["expires_in"];
	        this.user = this.convertValues(source["user"], Account);
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
	export class Submission {
	    id: number;
	    user_id: number;
	    question: string;
	    answer: string;
	    category: string;
	    tags: string[];
	    source: string;
	    remark: string;
	    status: string;
	    reviewer_note: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    reviewed_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Submission(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.source = source["source"];
	        this.remark = source["remark"];
	        this.status = source["status"];
	        this.reviewer_note = source["reviewer_note"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.reviewed_at = this.convertValues(source["reviewed_at"], null);
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
	export class SubmissionInput {
	    question: string;
	    answer: string;
	    category: string;
	    tags: string[];
	    source: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new SubmissionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.question = source["question"];
	        this.answer = source["answer"];
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.source = source["source"];
	        this.remark = source["remark"];
	    }
	}
	export class UserProfile {
	    id: number;
	    username: string;
	    nickname: string;
	
	    static createFrom(source: any = {}) {
	        return new UserProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	    }
	}

}

