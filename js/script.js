var DOMAIN                  = "DOMAIN_PLACEHOLDER",
    CROSSHAIR_HALF_SIZE     = 8,
    WIDTH                   = 2000,
    HEIGHT                  = 3000,
    MOUSEMOVE_DELAY         = 16,  // minimum time (in ms) between two mousemove events; 16ms ~= 60Hz
    MAX_NICKNAME_LENGTH     = 20,
    MAX_CHAT_MESSAGE_LENGTH = 256,
    this_script             = document.documentElement.lastChild,
    users                   = new Object(),
    mask_lines              = new Array(),
    my_last_x               = null,
    my_last_y               = null,
    my_use_pen              = 1,
    my_mouse_is_down        = null,
    chat_was_visible        = null,
    max_size                = 0,
    incoming_blobs          = [],
    is_decoding             = false,
    chat_div, myPicker, nickname_span, canvas, messages_div, mySocket,
    mask_canvas, ctx, mask_ctx, biggest_node, last_time, loading_box,
    progress_bar, warning_box, frame_div, frame


function distance(x1, y1, x2, y2){
    return Math.sqrt(Math.pow(x1 - x2, 2) + Math.pow(y1 - y2, 2))
}

function time_since_last_time(){
    var now, duration
    if (typeof(last_time) == "undefined"){
        last_time = (new Date()).getTime()
        return 0
    }
    now = (new Date()).getTime()
    duration = now - last_time
    if (duration >= MOUSEMOVE_DELAY){  // if duration is too short, act like if we were never called
        last_time = now
    }
    return duration
}

function destroy(){
    mySocket.close()
    for (var user_id in users) {
        // use hasOwnProperty to filter out keys from the Object.prototype
        if (users.hasOwnProperty(user_id)) {
            users[user_id].destroy()
        }
    }
    unwrap_document_from_iframe()
    delete window.webinvader_pad
}
this.destroy = destroy

//function init(){

    wrap_document_in_iframe()

    create_toolbar()
    create_chat_window()
    create_loading_box()
    create_warning_box()
    set_loading_on()
    create_socket()
//}
