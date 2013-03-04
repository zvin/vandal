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

