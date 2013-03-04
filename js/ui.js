function apply_default_style(element){
    element.style.color = "black"
}

function create_element(name){
    var element = document.createElement(name)
    apply_default_style(element)
    return element
}

function create_toolbar(){
    function create_button(background){
        var button = create_element("div")
        button.style.width      = "50px"
        button.style.height     = "50px"
        button.style.cssFloat   = "left"
        button.style.cursor     = "pointer"
        button.style.background = "url(http://" + DOMAIN + "/static/" + background + ") no-repeat scroll 6px 6px transparent"
        return button
    }
    function create_shadow(left){
        var shadow = create_element("div")
        shadow.style.position     = "absolute"
        shadow.style.width        = "38px"
        shadow.style.height       = "38px"
        shadow.style.margin       = "5px"
        shadow.style.cssFloat     = "left"
        shadow.style.top          = "0px"
        shadow.style.left         = left + "px"
        shadow.style.borderRadius = "5px"
        shadow.style.cursor       = "pointer"
        shadow.style.boxShadow    = "inset -1px 1px 3px 0px rgba(0, 0, 0, 0.3)"
    return shadow
    }
    toolbar = create_element("div")
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
    // toolbar
    toolbar.style.position = "fixed"
    toolbar.style.top  = "0px"
    toolbar.style.left = "0px"
    toolbar.style.width  = "269px"
    toolbar.style.height = "50px"
    toolbar.style.border = "2px solid #000000"
    toolbar.style.backgroundColor = "#FFFFFF"
    toolbar.style.borderRadius = "5px 0px 0px 5px"
    toolbar.style.boxShadow = "-2px 2px 5px 0px rgba(0, 0, 0, 0.3)"
    toolbar.id = "fixed_toolbar"
    // end toolbar
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
	loading_box = create_element("div")
    loading_box.style.position = "fixed"
    loading_box.style.top  = "50%"
    loading_box.style.left = "50%"
    loading_box.style.marginLeft = "-100px"
    loading_box.style.marginTop = "-25px"
    loading_box.style.width  = "200px"
    loading_box.style.height = "50px"
    loading_box.style.border = "2px solid #000000"
    loading_box.style.backgroundColor = "white"
    loading_box.style.padding = "10px"
    loading_box.style.fontFamily = "Arial, Helvetica, sans-serif"
    loading_box.style.fontSize = "36px"
    loading_box.style.fontWeight = "normal"
    loading_box.style.fontVariant = "normal"
    loading_box.style.fontStyle = "normal"
    loading_box.style.lineHeight = "50px"
    loading_box.style.textAlign = "center"
    loading_box.style.zIndex = reverse_zindex(0)
    progress_bar = document.createElement("progress")
    progress_bar.value = 0
    progress_bar.max = 100
    progress_bar.removeAttribute("value")
    progress_bar.style.width = "100%"
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
    var nickname_p = create_element("p"),
        choose_div = create_element("div"),
        icon_div   = create_element("div"),
        input      = create_element("input")
    chat_div = create_element("div")
    chat_div.style.position = "fixed"
    chat_div.style.top  = "0px"
    chat_div.style.right = "0px"
    chat_div.style.width  = "200px"
    chat_div.style.height = "100%"
    chat_div.style.border = "2px solid #000000"
    chat_div.style.backgroundColor = "white"
    chat_div.style.padding = "10px"
    chat_div.style.fontFamily = "Arial, Helvetica, sans-serif"
    chat_div.style.fontSize = "12px"
    chat_div.style.fontWeight = "normal"
    chat_div.style.fontVariant = "normal"
    chat_div.style.textAlign = "left"
    chat_div.style.fontStyle = "normal"
    chat_div.style.lineHeight = "16px"
    chat_div.style.zIndex = reverse_zindex(2)
    chat_div.style.overflowY = "auto"
    chat_div.style.overflowX = "hidden"
    chat_div.style.wordWrap = "break-word"
    chat_div.style.display = "block"
    nickname_p.style.margin = "0"
    nickname_p.style.marginTop = "10px"
    nickname_p.appendChild(document.createTextNode("Nickname : "))
    nickname_span = create_element("span")
    nickname_span.style.fontWeight = "bold"
    nickname_p.appendChild(nickname_span)
    chat_div.appendChild(nickname_p)
    choose_div.style.cssFloat = "left"
    choose_div.style.cursor = "pointer"
    choose_div.style.width = "45px"
    choose_div.style.padding = "5px"
    choose_div.style.margin = "10px 0"
    choose_div.style.border = "solid 1px #999"
    choose_div.style.borderRadius = "5px"
    choose_div.style.background = "linear-gradient(#ffffff 58%, #b2b2b2 98%)"
    choose_div.appendChild(document.createTextNode("Change"))
    choose_div.onclick = function(){
        send_change_nickname(prompt("Enter your new nickname (" + MAX_NICKNAME_LENGTH + " characters max):", ""))
    }
    chat_div.appendChild(choose_div)
    icon_div.style.cssFloat = "left"
    icon_div.style.position = "absolute"
    icon_div.style.top = "44px"
    icon_div.style.left = "150px"
    icon_div.style.height = "40px"
    icon_div.style.width = "60px"
    icon_div.style.background = "url(http://" + DOMAIN + "/static/buddy.png) no-repeat scroll 6px 6px transparent"
    chat_div.appendChild(icon_div)
    input.style.width = "100%"
    input.style.backgroundColor = "#f7f7f7"
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
    var p    = create_element("p"),
        span = create_element("span")
    p.style.marginTop = "10px"
    span.style.fontWeight = "bold"
    span.appendChild(document.createTextNode(username))
    p.appendChild(span)
    p.appendChild(document.createTextNode(" : " + msg))
    p.title = format_time(timestamp)
    add_message(p)
}

function add_chat_notification(msg, timestamp){
    var p = create_element("p")
    p.style.fontStyle = "italic"
    p.style.marginTop = "10px"
    p.appendChild(document.createTextNode(msg))
    p.title = format_time(timestamp)
    add_message(p)
}

function add_message(msg){
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
