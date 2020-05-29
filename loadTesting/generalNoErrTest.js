module.exports = function(params, options, client, callback) {
    let scenario = Math.round(Math.random()*2);
    let acc_1 = 1 + Math.floor(Math.random()*999);
    let acc_2 = 1 + Math.floor(Math.random()*999);
    while (acc_2 === acc_1) {
        acc_2 = 1 + Math.floor(Math.random()*999);
    }
    let unique_id = Math.round(Math.random()*100) * 1000000 + Math.ceil(Date.now() % 1000000);
    let amount = Math.round(Math.random()*200) / 100;

    switch (scenario) {
        case 0:
            options.path = '/ops/transfer'
            options.method = 'POST'
            options.body={
                fromId: acc_1,
                toId: acc_2,
                amount: amount
            }
            break;
        case 1:
            options.path = '/accounts/' + acc_1
            options.method = 'GET'
            options.body={}
            break;
        case 2:
            options.path = '/accounts/'
            options.method = 'POST'
            options.body={
                id: unique_id,
                balance: 100
            }
            break;
        default:
            console.log(`Sorry, we are out of ${scenario}.`);
    }

    let request = client(options, callback);
    let jsonBody = JSON.stringify(options.body)
    request.write(jsonBody)

    return request;
}