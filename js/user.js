/**
 * @constructor
 */
function User(user_id, position, color, mouse_is_down, it_is_me, nickname, use_pen){
    this.user_id       = user_id
    this.position      = position
    this.color         = color
    this.mouse_is_down = mouse_is_down
    this.it_is_me      = it_is_me
    this.nickname      = nickname
    this.use_pen       = use_pen
    users[this.user_id] = this
    this.create_label()
}

User.prototype.reposition_label = function(){
    if (this.it_is_me) return
    if (this.label.style.transform === undefined){
        this.label.style.webkitTransform = "translate(" + (this.position[0] - CROSSHAIR_HALF_SIZE) + "px, " + (this.position[1] - CROSSHAIR_HALF_SIZE) + "px)"
    }else{
        this.label.style.transform = "translate(" + (this.position[0] - CROSSHAIR_HALF_SIZE) + "px, " + (this.position[1] - CROSSHAIR_HALF_SIZE) + "px)"
    }
}

User.prototype.create_label = function(){
    var text = document.createTextNode(this.nickname || this.user_id)
    if (this.it_is_me){
        nickname_span.appendChild(text)
    }else{
        this.label = create_element("div")
        this.label.appendChild(text)
        this.label.style.position = "absolute"
        this.label.style.left = 0
        this.label.style.top = 0
        this.label.style.paddingLeft = "16px"
        this.label.style.zIndex = 1
        this.label.style.background = "url(http://" + DOMAIN + ":" + HTTP_PORT + "/static/crosshair.png) no-repeat"
        this.reposition_label()
        frame_div.appendChild(this.label)
    }
}

User.prototype.mouse_up = function(){
    this.mouse_is_down = false
}

User.prototype.mouse_down = function(){
    this.mouse_is_down = true
}

User.prototype.mouse_move = function(x, y, duration){
    if (this.mouse_is_down){
        lines_to_draw.push([
            this.position[0], this.position[1],          // origin
            x, y,                                        // destination
            duration,                                    // duration
            this.color[0], this.color[1], this.color[2], // color
            this.use_pen,                                // pen or eraser
            ctx                                          // context
        ])
        if (this.use_pen && this.it_is_me){
            mask_shift()
        }
    }
    this.position = [x, y]
}

User.prototype.change_color = function(red, green, blue){
    this.color = [red, green, blue]
}

User.prototype.change_tool = function(use_pen){
    this.use_pen = use_pen
}

User.prototype.change_nickname = function(nickname, timestamp){
    add_chat_notification(this.get_label() + " is now known as " + nickname, timestamp)
    this.nickname = nickname
    if (this.it_is_me){
        nickname_span.removeChild(nickname_span.firstChild)
        nickname_span.appendChild(document.createTextNode(nickname))
    }else{
        this.label.removeChild(this.label.firstChild)
        var text = document.createTextNode(this.nickname)
        this.label.appendChild(text)
    }
}

User.prototype.destroy = function(){
    if (this.label){
        frame_div.removeChild(this.label)
    }
    delete users[this.user_id]
}

User.prototype.get_label = function(){
    return this.nickname || this.user_id
}
