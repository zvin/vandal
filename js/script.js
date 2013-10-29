var DOMAIN                  = "DOMAIN_PLACEHOLDER",
    HTTPS_PORT              = "HTTPS_PORT_PLACEHOLDER",
    HTTP_PORT               = "HTTP_PORT_PLACEHOLDER",
    SECURE                  = (document.location.protocol == "https:") ? "s" : "",
    CROSSHAIR_HALF_SIZE     = 8,
    WIDTH                   = 2000,
    HEIGHT                  = 3000,
    MOUSEMOVE_DELAY         = 16,  // minimum time (in ms) between two mousemove events; 16ms ~= 60Hz
    MAX_NICKNAME_LENGTH     = 20,
    MAX_CHAT_MESSAGE_LENGTH = 256,
    ZOOM_MIN                = 0.5,
    ZOOM_MAX                = 10,
    ZOOM_FACTOR             = 1.1,
    users                   = new Object(),
    mask_lines              = new Array(),
    lines_to_draw           = new Array(),
    my_last_x               = null,
    my_last_y               = null,
    my_use_pen              = 1,
    my_mouse_is_down        = null,
    chat_was_visible        = null,
    zoom                    = 1.0,
    mask_dirty              = false,
    chat_div, myPicker, nickname_span, canvas, messages_div, mySocket,
    mask_canvas, ctx, mask_ctx, last_time, loading_box,
    progress_bar, warning_box, frame_div, frame, error_message_div, zoom_slider


function distance(x1, y1, x2, y2){
    return Math.sqrt(Math.pow(x1 - x2, 2) + Math.pow(y1 - y2, 2))
}

function logn(i, base){
    return Math.log(i) / Math.log(base)
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

function remove_all_users(){
    for (var user_id in users) {
        // use hasOwnProperty to filter out keys from the Object.prototype
        if (users.hasOwnProperty(user_id)) {
            users[user_id].destroy()
            delete users[user_id]
        }
    }
}

function destroy(){
    mySocket.close()
    remove_all_users()
    unwrap_document_from_iframe()
    delete window.webinvader_pad
}
this.destroy = destroy

var requestAnimFrame = window.requestAnimationFrame || window.webkitRequestAnimationFrame

function render_loop(){
    requestAnimFrame(render_loop)
    for (var user_id in users) {
        // use hasOwnProperty to filter out keys from the Object.prototype
        if (users.hasOwnProperty(user_id)) {
            users[user_id].reposition_label()
        }
    }
    while(lines_to_draw.length > 0){
        draw_line.apply(this, lines_to_draw.shift())
    }
    // anti-lag mask canvas
    if (mask_dirty){
        mask_ctx.clearRect(0, 0, mask_canvas.width, mask_canvas.height)
        for (var i=0; i < mask_lines.length; i++){
            draw_line.apply(this, mask_lines[i])
        }
        mask_dirty = false
    }
}
render_loop()

//function init(){

    wrap_document_in_iframe()

    create_toolbar()
    create_chat_window()
    create_loading_box()
    create_warning_box()
    set_loading_on()
    create_socket()
    create_canvas()

//}
