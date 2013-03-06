function apply_default_style(element){
    element.style.color = "black"
}

function create_element(name, style){
    var element = document.createElement(name)
    apply_default_style(element)
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
    button_close.onclick = function(){
        if (confirm("Do you really want to close the draw pad?")){
            destroy()
        }
    }
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
    toolbar.style.zIndex = reverse_zindex(1)
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
        "z-index"         :  reverse_zindex(0)
    })
    progress_bar = document.createElement("progress")
    progress_bar.value = 0
    progress_bar.max = 100
    progress_bar.removeAttribute("value")
    progress_bar.style.setProperty("width", "100%")
    loading_box.appendChild(progress_bar)
    document.body.appendChild(loading_box)
}

function set_loading_on(){
	loading_box.style.display = "block"
}

function set_loading_off(){
	loading_box.style.display = "none"
}

function create_chat_window(){
    var nickname_p = create_element("p", {
        "margin"    : "0",
        "margin-top": "10px"
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
        "z-index"         : reverse_zindex(2),
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
    var p = create_element("p", {"margin-top": "10px"})
    var span = create_element("span", {"font-weight": "bold"})
    span.appendChild(document.createTextNode(username))
    p.appendChild(span)
    p.appendChild(document.createTextNode(" : " + msg))
    add_message(p, timestamp)
}

function add_chat_notification(msg, timestamp){
    var p = create_element("p", {"font-style": "italic", "margin-top": "10px"})
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

function put_embeds_down(){
    // dirty hack to be able to draw over youtube flash videos
    var embeds = document.getElementsByTagName("embed")
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

function reverse_zindex(level){
    // level 0 is on top
    return MAX_ZINDEX - level
}

function decrease_zindexes(element, levels, limit){
    // Walks the element's childs and decreases all zindexes by 'levels' value
    // if the value is superior or equal to limit.
    // This is useful on websites that use the maximum zindex 2147483647
    if (element.nodeType == document.ELEMENT_NODE){
        var style = document.defaultView.getComputedStyle(element, null)
        if ((style.zIndex != "") && (style.zIndex != "auto") && (style.zIndex >= limit)){
            console.log(element, style.zIndex)
            element.style.zIndex = style.zIndex - levels
        }
        for (var i=0; i<element.childNodes.length; i++){
            decrease_zindexes(element.childNodes[i], levels, limit)
        }
    }
}

function get_my_color(){
    return [
        Math.round(myPicker.rgb[0] * 255),
        Math.round(myPicker.rgb[1] * 255),
        Math.round(myPicker.rgb[2] * 255)
    ]
}


function show_area(node){
//    console.log(node, node.clientWidth * node.clientHeight)
//    if ((typeof(canvas) != "undefined") && (node == canvas)){
//        return 0
//    }
    return node.clientWidth * node.clientHeight
}

function walk_dom_max(node, func){
    var max_value = 0,
        max_node  = null,
        value, node_and_value, i

    for (i=0; i<node.childNodes.length; i++){
        if (node.childNodes[i].nodeType == Node.ELEMENT_NODE){
            value = func(node.childNodes[i])
            if (value > max_value){
                max_value = value
                max_node = node.childNodes[i]
            }
            node_and_value = walk_dom_max(node.childNodes[i], func)
            if (node_and_value[1] > max_value){
                max_value = node_and_value[1]
                max_node  = node_and_value[0]
            }
        }
    }

    return [max_node, max_value]
}

function hop(node){
    var left  = node.offsetLeft,
        width = node.offsetWidth,
        right = 0
    // next line won't work in chromium
//    var right = Math.max(node.offsetParent.offsetWidth, node.scrollWidth) - (left + width)
//    console.log(
//        node.offsetLeft,
//        max(node.offsetParent.offsetWidth, node.scrollWidth) - (node.offsetLeft + node.offsetWidth)
//    )
    return [left, width, right]
}

function window_width(){
    var docElemProp = window.document.documentElement["clientWidth"];
    return window.document.compatMode === "CSS1Compat" && docElemProp || window.document.body["clientWidth"] || docElemProp;
}

function there_is_a_scrollbar(){
    //$(document).width() > $(window).width()
//    return (document.getElementsByTagName("html")[0].scrollWidth > document.body.scrollWidth)
    return (document.getElementsByTagName("html")[0].scrollWidth > window_width())
}

function hide_everything(){
    canvas.style.display = "none"
    mask_canvas.style.display = "none"
    toolbar.style.display = "none"
    chat_was_visible = (chat_div.style.display != "none")
    chat_div.style.display = "none"
}

function unhide_everything(){
    canvas.style.display = null
    mask_canvas.style.display = null
    toolbar.style.display = null
    if (chat_was_visible){
        chat_div.style.display = null
    }
}

function reposition_canvas(){
    hide_everything()
    if (typeof(biggest_node) == "undefined"){
        biggest_node = walk_dom_max(document.body, show_area)[0]
//        console.log(biggest_node.offsetWidth, biggest_node.style.width)
//        biggest_node.style.width = window.getComputedStyle(biggest_node).width
    }
//    var biggest_node = walk_dom_max(document.body, show_area)[0]
    if (biggest_node == null){
//        console.log(Math.round(window_width() / 2) - (WIDTH / 2))
        canvas_set_left( Math.round(window_width() / 2) - (WIDTH / 2) )
        unhide_everything()
        return
    }
//    console.log(biggest_node, biggest_node.style.display, biggest_node.offsetWidth)
//    console.log(hop(biggest_node))
    var left_width_right = hop(biggest_node),
        left             = left_width_right[0],
        width            = left_width_right[1],
        right            = left_width_right[2]
//    document.body.removeChild(canvas)
    if (there_is_a_scrollbar()){
        canvas_set_left( Math.round(document.getElementsByTagName("html")[0].scrollWidth / 2) - (WIDTH / 2) )
    }else{
        canvas_set_left( left + Math.round(width / 2) - (WIDTH / 2) )
    }
//    document.body.appendChild(canvas)
    unhide_everything()
}


function canvas_set_left(offset){
    canvas.style.left = offset + "px"
    mask_canvas.style.left = offset + "px"
}

function mask_redraw(){
    if (mask_lines.length > max_size){
//        console.log(mask_lines.length)
        max_size = mask_lines.length
    }
    mask_ctx.clearRect(0, 0, mask_canvas.width, mask_canvas.height)
    for (var i=0; i < mask_lines.length; i++){
        draw_line.apply(this, mask_lines[i].concat([mask_ctx]))
    }
}

function mask_push(line){
    //console.log("push")
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
//        console.log("draw delta", lines)
        for(var i=0; i<lines.length; i++){
            draw_line.apply(this, lines[i].concat([ctx]))
        }
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
    for(var i=0; i<messages.length; i++) {
        if (messages[i][0] == "") {
            add_chat_notification(messages[i][1], messages[i][2])
        } else {
            add_chat_message(messages[i][0], messages[i][1], messages[i][2])
        }
    }
}