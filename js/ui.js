function create_element(name, style){
    var element = document.createElement(name)
    for (var key in style) {
        if (style.hasOwnProperty(key)) {
            element.style.setProperty(key, style[key])
        }
    }
    return element
}

function create_toolbar(){
    function create_button(background){
        var button = create_element("div", {
            "width"     : "50px",
            "height"    : "50px",
            "float"     : "left",
            "cursor"    : "pointer",
            "background": "url(http://" + DOMAIN + "/static/" + background + ") no-repeat scroll 6px 6px transparent"
        })
        return button
    }
    function create_shadow(left){
        return create_element("div", {
            "position"      : "absolute",
            "width"         : "38px",
            "height"        : "38px",
            "margin"        : "5px",
            "float"         : "left",
            "top"           : "0px",
            "left"          : left + "px",
            "border-radius" : "5px",
            "cursor"        : "pointer",
            "box-shadow"    : "inset -1px 1px 3px 0px rgba(0, 0, 0, 0.3)"
        })
    }
    toolbar = create_element("div", {
        "position"        : "fixed",
        "top"             : "0px",
        "left"            : "0px",
        "width"           : "269px",
        "height"          : "50px",
        "border"          : "2px solid #000000",
        "background-color": "#FFFFFF",
        "border-radius"   : "5px 0px 0px 5px",
        "box-shadow"      : "-2px 2px 5px 0px rgba(0, 0, 0, 0.3)"
    })
    toolbar.id = "fixed_toolbar"  // needed for jscolor
    var button_toggle_chat = create_button("chat.png"),
        button_color       = create_element("input"),
        handle             = create_element("div"),
        tool_on            = create_shadow(49),
        chat_on            = create_shadow(150),
        button_pen         = create_button("pen.png"),
        button_eraser      = create_button("eraser.png"),
        button_close       = create_button("close.png"),
        click_x            = null, // toolbar movements
        click_y            = null  // toolbar movements
    button_pen.onclick = function(){
        send_change_tool(1)
        tool_on.style.left = "49px"
    }
    button_eraser.onclick = function(){
        send_change_tool(0)
        tool_on.style.left = "100px"
    }
    button_close.onclick = destroy
    // color
    if (document.compatMode == "BackCompat"){
        button_color.style.width = "26px"
        button_color.style.height = "28px"
    }else{
        button_color.style.width = "24px"
        button_color.style.height = "24px"
    }
    button_color.style.cssFloat = "left"
    button_color.style.cursor = "pointer"
    button_color.style.margin = "12px"
    button_color.style.border = "1px solid black"
    button_color.style.padding = "0px"
    myPicker = new jscolor.color(button_color, {})
    myPicker.fromString('000000')
    button_color.onchange = function(){
        send_change_color.apply(this, get_my_color())
    }
    // end color
    // toggle_chat
    function toggle_chat_window(){
        if (chat_div.style.display != "none"){
            chat_div.style.display = "none"
            chat_on.style.display = "none"
        }else{
            chat_div.style.display = "block"
            chat_on.style.display = "block"
        }
    }
    button_toggle_chat.onclick = toggle_chat_window
    chat_on.onclick = toggle_chat_window
    // end toggle_chat
    // handle
    handle.style.width = "14px"
    handle.style.height = "50px"
    handle.style.cssFloat = "right"
    handle.style.background = "url(http://" + DOMAIN + "/static/handle.png) no-repeat scroll 0 0 transparent"
    // end handle
    toolbar.appendChild(button_color)
    toolbar.appendChild(button_pen)
    toolbar.appendChild(button_eraser)
    toolbar.appendChild(chat_on)
    toolbar.appendChild(button_toggle_chat)
    toolbar.appendChild(tool_on)
    toolbar.appendChild(button_close)
    toolbar.appendChild(handle)
    toolbar.style.zIndex = 4
    document.body.appendChild(toolbar)


    function mousemove(ev){
        toolbar.style.left = ev.clientX - click_x + "px"
        toolbar.style.top  = ev.clientY - click_y + "px"
    }

    handle.onmousedown = function(ev){
        document.body.style.cursor = "move"
        mask_canvas.style.cursor = "move"
        click_x = ev.clientX - parseInt(toolbar.style.left, 10)
        click_y = ev.clientY - parseInt(toolbar.style.top, 10)
        document.addEventListener("mousemove", mousemove, false)
        return false
    }

    document.addEventListener(
        "mouseup",
        function(ev){
            document.body.style.cursor = ""
            mask_canvas.style.cursor = "crosshair"
            document.removeEventListener("mousemove", mousemove, false)
        },
        false
    )
}

function create_loading_box(){
	loading_box = create_element("div", {
        "position"        : "fixed",
        "top"             : "50%",
        "left"            : "50%",
        "margin-left"     : "-100px",
        "margin-top"      : "-25px",
        "width"           : "200px",
        "height"          : "50px",
        "border"          : "2px solid #000000",
        "background-color": "white",
        "padding"         : "10px",
        "font-family"     : "Arial, Helvetica, sans-serif",
        "font-size"       : "36px",
        "font-weight"     : "normal",
        "font-variant"    : "normal",
        "font-style"      : "normal",
        "line-height"     : "50px",
        "text-align"      : "center",
        "z-index"         :  5
    })
    progress_bar = document.createElement("progress")
    progress_bar.value = 0
    progress_bar.max = 100
    progress_bar.removeAttribute("value")
    progress_bar.style.setProperty("width", "100%")
    loading_box.appendChild(progress_bar)
    document.body.appendChild(loading_box)
}

function create_warning_box(){
	warning_box = create_element("div", {
        "position"        : "fixed",
        "top"             : "50%",
        "left"            : "50%",
        "margin-left"     : "-150px",
        "margin-top"      : "-25px",
        "width"           : "300px",
        "height"          : "50px",
        "border"          : "6px solid red",
        "background-color": "white",
        "color"           : "red",
        "padding"         : "10px",
        "font-family"     : "Arial, Helvetica, sans-serif",
        "font-size"       : "36px",
        "font-weight"     : "normal",
        "font-variant"    : "normal",
        "font-style"      : "normal",
        "line-height"     : "50px",
        "text-align"      : "center",
        "display"         : "none",
        "z-index"         :  5
    })
    warning_box.appendChild(document.createTextNode("Disconnected"))
    document.body.appendChild(warning_box)
}

function set_loading_on(){
	loading_box.style.display = "block"
}

function set_loading_off(){
	loading_box.style.display = "none"
}

function create_chat_window(){
    var nickname_p = create_element("p", {
        "margin"     : "0",
        "margin-top" : "10px",
        "text-indent": 0
    })
    var choose_div = create_element("div", {
        "float"        : "left",
        "cursor"       : "pointer",
        "width"        : "45px",
        "padding"      : "5px",
        "margin"       : "10px 0",
        "border"       : "solid 1px #999",
        "border-radius": "5px",
        "background"   : "linear-gradient(#ffffff 58%, #b2b2b2 98%)"
    })
    var icon_div = create_element("div", {
        "float"     : "left",
        "position"  : "absolute",
        "top"       : "44px",
        "left"      : "150px",
        "height"    : "40px",
        "width"     : "60px",
        "background": "url(http://" + DOMAIN + "/static/buddy.png) no-repeat scroll 6px 6px transparent"
    })
    var input = create_element("input", {
        "width"           : "100%",
        "background-color": "#f7f7f7"
    })
    chat_div = create_element("div", {
        "position"        : "fixed",
        "top"             : "0px",
        "right"           : "0px",
        "width"           : "200px",
        "height"          : "100%",
        "border"          : "2px solid #000000",
        "background-color": "white",
        "padding"         : "10px",
        "font-family"     : "Arial, Helvetica, sans-serif",
        "font-size"       : "12px",
        "font-weight"     : "normal",
        "font-variant"    : "normal",
        "text-align"      : "left",
        "font-style"      : "normal",
        "line-height"     : "16px",
        "z-index"         : 3,
        "overflow-y"      : "auto",
        "overflow-x"      : "hidden",
        "word-wrap"       : "break-word",
        "display"         : "block"
    })
    nickname_p.appendChild(document.createTextNode("Nickname : "))
    nickname_span = create_element("span", {"font-weight": "bold"})
    nickname_p.appendChild(nickname_span)
    chat_div.appendChild(nickname_p)
    choose_div.appendChild(document.createTextNode("Change"))
    choose_div.onclick = function(){
        send_change_nickname(prompt("Enter your new nickname (" + MAX_NICKNAME_LENGTH + " characters max):", ""))
    }
    chat_div.appendChild(choose_div)
    chat_div.appendChild(icon_div)
    input.onkeyup = function(event){
        if (event.keyCode == 13){
            send_chat_message(input.value)
            input.value = ""
        }
    }
    chat_div.appendChild(input)
    messages_div = create_element("div")
    chat_div.appendChild(messages_div)
    document.body.appendChild(chat_div)
}

function format_time(timestamp){
    return (new Date(timestamp * 1000)).toLocaleString()
}

function add_chat_message(username, msg, timestamp){
    var p = create_element("p", {"margin-top": "10px", "text-indent": 0})
    var span = create_element("span", {"font-weight": "bold"})
    span.appendChild(document.createTextNode(username))
    p.appendChild(span)
    p.appendChild(document.createTextNode(" : " + msg))
    add_message(p, timestamp)
}

function add_chat_notification(msg, timestamp){
    var p = create_element(
        "p", {"font-style": "italic", "margin-top": "10px", "text-indent": 0}
    )
    p.appendChild(document.createTextNode(msg))
    add_message(p, timestamp)
}

function add_message(msg, timestamp){
    msg.title = format_time(timestamp)
    if (messages_div.firstChild == null){
        messages_div.appendChild(msg)
    }else{
        messages_div.insertBefore(msg, messages_div.firstChild)
    }
}

function get_my_color(){
    return myPicker.rgb.map(function(x){return Math.round(x * 255)})
}

function mask_redraw(){
    if (mask_lines.length > max_size){
        max_size = mask_lines.length
    }
    mask_ctx.clearRect(0, 0, mask_canvas.width, mask_canvas.height)
    for (var i=0; i < mask_lines.length; i++){
        draw_line.apply(this, mask_lines[i].concat([mask_ctx]))
    }
}

function mask_push(line){
    mask_lines.push(line)
    draw_line.apply(this, line.concat([mask_ctx]))
}

//cpt = 0
function mask_shift(){
    mask_lines.shift()
    //if (cpt % 100 == 0){
        mask_redraw()
    //}
    //cpt++
}

function draw_line(x1, y1, x2, y2, duration, red, green, blue, use_pen, context){
    if (duration <= 0){
        return
    }
    var d = distance(x1, y1, x2, y2)
    if (d <= 0){
        return
    }
    var speed = d / duration
    if (!use_pen){ // not use_pen: use eraser
        context.globalCompositeOperation = "destination-out"
    }else{
        context.globalCompositeOperation = "source-over"
        context.strokeStyle = "rgb(" + red + "," + green + "," + blue + ")"
    }
    context.lineWidth = 1 / (1.3 + (3 * speed))
    context.beginPath()
    context.moveTo(x1, y1)
    context.lineTo(x2, y2)
    context.stroke()
    context.closePath()
}

function copy_img_in_canvas(blob_id){
    var img = new Image()
    img.src = blob_id;
    img.onload = function(){
        ctx.drawImage(img, 0, 0)
    }
}

function draw_delta(lines){
    if (lines) {
        lines.forEach(function(line){draw_line.apply(this, line.concat([ctx]))})
    }
}

function load_image(url){
    var request = new XMLHttpRequest();
    request.onprogress = updateProgressBar;
    request.onload = showImage;
    request.onloadend = set_loading_off;
    request.open("GET", url, true);
    request.responseType = 'arraybuffer';
    request.send(null);

    function updateProgressBar(e){
        if (e.lengthComputable)
            progress_bar.value = e.loaded / e.total * 100;
    }

    function showImage(){
        var blob = new Blob([new Uint8Array(request.response)], {"type": "image/png"});
        copy_img_in_canvas(URL.createObjectURL(blob))
    }
}

function display_chat_log(messages) {
    if (messages == null) return
    messages.forEach(function(message){
        if (message[0] == "") {
            add_chat_notification(message[1], message[2])
        } else {
            add_chat_message(message[0], message[1], message[2])
        }
    })
}

function put_embeds_down(){
    // dirty hack to be able to draw over youtube flash videos
    var embeds = frame.contentWindow.document.getElementsByTagName("embed")
    for (var i=0; i<embeds.length; i++){
        var embed = embeds[i]
        if ((embed.getAttribute("type") == "application/x-shockwave-flash") && (embed.getAttribute("wmode") == null)) {
            var parent = embed.parentNode
            embed.setAttribute("wmode", "opaque")
            parent.removeChild(embed)
            setTimeout(function(){parent.appendChild(embed)}, 0)
            setTimeout(function(){embed.stopVideo()}, 1000)
        }
    }
}


function unwrap_document_from_iframe(){
    document.documentElement.innerHTML = frame.contentWindow.document.documentElement.innerHTML
}

function wrap_document_in_iframe(){
    var height = document.documentElement.scrollHeight
    var content = []
    for (var i=0; i<document.documentElement.childNodes.length; i++){
        var child = document.documentElement.childNodes[i]
        content.push(document.documentElement.removeChild(child))
    }
    document.documentElement.innerHTML = ""
    frame_div = create_element("div", {
        "position": "relative",
        "width"   : WIDTH + "px",
        "margin"  : "auto",
    })
    frame = create_element("iframe", {
        "width"   : "100%",
        "height"  : height + "px",
        "overflow": "hidden",
        "border"  : 0
    })
    frame.onload = put_embeds_down
    frame_div.appendChild(frame)
    document.body.appendChild(frame_div)
    setTimeout(
        function(){
            // copy html content
            for (var i=0; i<content.length; i++){
                frame.contentWindow.document.documentElement.appendChild(content.pop())
            }
            // copy globals
            for (var key in window){
                if (window.hasOwnProperty(key)){
                    try{
                        frame.contentWindow[key] = window[key]
                    }catch(err){
                    }
                }
            }
        },
        100
    )
}
