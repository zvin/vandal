var EventType = {
	join            : 1,
	leave           : 2,
	welcome         : 3,
	change_color    : 4,
	change_tool     : 5,
	mouse_move      : 6,
	mouse_up        : 7,
	mouse_down      : 8,
	chat_message    : 9,
	change_nickname : 10,
	error           : 11
}

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
    canvas = create_element("canvas", {
        "position": "absolute",
        "top"     : document.body.clientTop,
        "z-index" : reverse_zindex(5),
        "outline" : "10px dashed #F30B55"
    })
    mask_canvas = create_element("canvas", {
        "position": "absolute",
        "cursor"  : "crosshair",
        "top"     : document.body.clientTop,
        "z-index" : reverse_zindex(3)
    })
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

function mousemove(ev){
//    console.log(ev)
    var duration = time_since_last_time()
    if (duration < MOUSEMOVE_DELAY) return
    // offset for chrome and layer for mozilla
    var mouse_x  = ev.offsetX || ev.layerX,
        mouse_y  = ev.offsetY || ev.layerY,
        my_color
    send_event([
        EventType.mouse_move,
        mouse_x,               // x
        mouse_y,               // y
        duration               // duration
    ])
    if (my_use_pen && my_mouse_is_down && my_last_x != null){
        // anti-lag system
        my_color = get_my_color()
        mask_push([
            my_last_x, my_last_y,                              // origin
            mouse_x, mouse_y,                                  // destination
            duration,                                          // duration
            my_color[0], my_color[1], my_color[2],             // color
            true                                               // pen not eraser
        ])
    }
    my_last_x = mouse_x
    my_last_y = mouse_y
}

function mousedown(ev){
    send_event([EventType.mouse_down])
    my_mouse_is_down = true
    return false // prevent text selection in chrome
}

function mouseup(ev){
    send_event([EventType.mouse_up])
    my_mouse_is_down = false
    ev.stopPropagation()
}

function gotmessage(event){
    var type = event.shift(),            // event_type
        user_id, user
    if (type == EventType.error){
        alert(event[0])
        document.location.reload()
    }
    if (type != EventType.welcome){
        user_id = event.shift()
        user = users[user_id]
    }
    if (type == EventType.join){
        new User(
            user_id,
            event[0], // position
            event[1], // color
            event[2], // mouse_is_down
            event[3], // you
            event[4], // nickname
            event[5]  // use_pen
        )
        if (event[6] != 0) {
            add_chat_notification("user " + event[4] + " joined ", event[6])
        }
    }else if (type == EventType.leave){
        add_chat_notification("user " + user.nickname + " left ", event[0])
        user.destroy()
    }else if (type == EventType.mouse_up){
        user.mouse_up()
    }else if (type == EventType.mouse_down){
        user.mouse_down()
    }else if (type == EventType.mouse_move){
        user.mouse_move.apply(user, event)
    }else if (type == EventType.change_color){
        user.change_color.apply(user, event)
    }else if (type == EventType.change_tool){
        user.change_tool.apply(user, event)
    }else if (type == EventType.chat_message){
        add_chat_message(user.get_label(), event[0], event[1])
    }else if (type == EventType.welcome){
        load_image("http://" + DOMAIN + "/" + event[0])          // image url
        draw_delta(event[1])                                     // delta
        display_chat_log(event[2])                               // chat history
    }else if (type == EventType.change_nickname){
        user.change_nickname.apply(user, event)
    }
}

