class InstantJasmineReporter {

    constructor() {}

    suiteStarted(result) {
        console.log("########################", result.id, result.description, "########################");
    };

    specStarted(result) {
        console.log("##########", result.id, result.description, "##########");
    };

}

module.exports = {
    InstantJasmineReporter
}
