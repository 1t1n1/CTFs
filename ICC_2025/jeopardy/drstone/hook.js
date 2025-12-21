var module = Process.getModuleByName("libdrstone.so")
var a = module.findExportByName("_cgoexp_bc91ffcaf021_Java_icc2025_drstone_DrStone_click")
Interceptor.attach(a, {
    onEnter: function(args) {
			console.log("a")
	        // console.log("[*] target_function called");
	        // console.log("    arg0: " + args[0].toInt32());
	    },
    onLeave: function(retval) {
	        // console.log("    return value: " + retval.toInt32());
			console.log("a")
	    }
});


// var module = Process.getModuleByName("libdrstone.so")
// var symbols = module.enumerateExports();
// console.log(symbols.length)
// symbols.forEach((a) => {
// 	console.log(a.name)
// })
