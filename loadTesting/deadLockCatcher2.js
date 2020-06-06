let counter = 0

module.exports = function(params, options, client, callback) {
    options.path = '/ops/transfer'
    options.method = 'POST'
    const direction = counter % 2
    if (direction) {
        options.body = {
            fromId: 1,
            toId: 2,
            amount: 10
        }
        // console.log('A')
    } else {
        options.body = {
            fromId: 2,
            toId: 1,
            amount: 100
        }
        // console.log('B')
    }

    let request = client(options, callback);
    let jsonBody = JSON.stringify(options.body)
    request.write(jsonBody)

    counter++;
    return request;
}