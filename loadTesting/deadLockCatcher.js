module.exports = function(params, options, client, callback) {
    options.path = '/ops/transfer'
    options.method = 'POST'
    let testIds = [55, 66]
    let indx = Math.round(Math.random());
    let fromId = testIds[indx]
    let toId = testIds[1 - indx]
    options.body={
        fromId,
        toId,
        amount: 0.01
    }

    let request = client(options, callback);
    let jsonBody = JSON.stringify(options.body)
    request.write(jsonBody)

    return request;
}