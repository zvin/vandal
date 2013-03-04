function send_event(event){
//    console.log(event, msgpack.pack(event, true).length)
//    console.log(msgpack.unpack(msgpack.pack(event, true)), event)
    if (mySocket.readyState == 1){
//        mySocket.send(JSON.stringify(event))
        var data = msgpack.pack(event)
        var array = new Uint8Array(data.length)
        for (var i = 0; i < data.length; i++) {
          array[i] = data[i]
        }
        mySocket.send(array.buffer)
    }
}

function send_change_color(red, green, blue){
    send_event([
        EventType.change_color,
        red,
        green,
        blue
    ])
}

function send_change_tool(use_pen){
    send_event([
        EventType.change_tool,
        use_pen
    ])
    my_use_pen = use_pen
}

function send_change_nickname(nickname){
    if ((nickname != null) && (nickname.length <= MAX_NICKNAME_LENGTH)){
        send_event([
            EventType.change_nickname,
            nickname
        ])
    }
}

function send_chat_message(msg){
    if (msg != ""){
        send_event([
            EventType.chat_message,
            msg
        ])
    }
}

function decode_msgpack(data){
    // Reading a blob is an async operation but we want to keep the messages in
    // order, so we have a queue of incoming messages and we decode them one at
    // a time.
    function decode(){
        is_decoding = true
        var reader = new FileReader()
        reader.onloadend = function(evt) {
            if (evt.target.readyState == FileReader.DONE) {
                var bytes = new Uint8Array(evt.target.result)
                gotmessage(msgpack.unpack(bytes))
                if (incoming_blobs.length > 0){
                    decode()
                } else {
                    is_decoding = false
                }
            }
        }
        reader.readAsArrayBuffer(incoming_blobs.shift());
    }
    incoming_blobs.push(data)
    if (!is_decoding){
        decode()
    }
}

function create_socket(){
    mySocket = new WebSocket(
        'ws://' + DOMAIN + "/ws?u=" + encodeURIComponent(
            document.location.href
        )
    )
    mySocket.onmessage = function(e){decode_msgpack(e.data)}
    mySocket.onclose = function(){
        alert("Connection to the server closed, please reload.")
        destroy()
    }
    canvas = create_element("canvas")
    mask_canvas = create_element("canvas")
    canvas.style.position = "absolute"
    mask_canvas.style.position = "absolute"
    mask_canvas.style.cursor = "crosshair"
    canvas.style.top = document.body.clientTop
    mask_canvas.style.top = document.body.clientTop
    canvas.style.zIndex = reverse_zindex(5)
    mask_canvas.style.zIndex = reverse_zindex(3)
    canvas.width = WIDTH
    mask_canvas.width = WIDTH
    canvas.height = HEIGHT
    mask_canvas.height = HEIGHT
    document.body.appendChild(canvas)
    document.body.appendChild(mask_canvas)
    ctx = canvas.getContext("2d")
    mask_ctx = mask_canvas.getContext("2d")
    ctx.strokeStyle = "rgb(0,0,0)"
    mask_canvas.onmousedown = mousedown
    mask_canvas.onmouseup   = mouseup
    mask_canvas.onmousemove = mousemove
    mask_canvas.onmouseout  = mouseup
//    window.setInterval(mask_redraw, 30)
}

