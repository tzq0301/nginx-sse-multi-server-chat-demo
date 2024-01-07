const EventSource = require('eventsource')

const source = new EventSource("http://127.0.0.1/group/1234")

source.onmessage = event => {
    console.log(event.data)
}

process.on('SIGTERM', function () {
    source.close()
    process.exit(0)
});
