/**
 * @constructor
 */
function User(user_id, position, color, mouse_is_down, it_is_me, nickname, use_pen){

    this.reposition_label = function(){
        if (this.it_is_me) return
        this.label.style.left = (canvas.offsetLeft + this.position[0] - CROSSHAIR_HALF_SIZE) + "px"
        this.label.style.top = (this.position[1] - CROSSHAIR_HALF_SIZE) + "px"
    }

    this.create_label = function(){
        var text = document.createTextNode(this.nickname || this.user_id)
        if (this.it_is_me){
            nickname_span.appendChild(text)
        }else{
            this.label = create_element("div")
            this.label.appendChild(text)
            this.label.style.position = "absolute"
            this.label.style.paddingLeft = "16px"
            this.label.style.zIndex = reverse_zindex(4)
            this.label.style.background = "url(http://" + DOMAIN + "/static/crosshair.png) no-repeat"
            this.reposition_label()
            frame_div.appendChild(this.label)
        }
    }

    this.init = function(user_id, position, color, mouse_is_down, it_is_me, nickname, use_pen){
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

    this.init(user_id, position, color, mouse_is_down, it_is_me, nickname, use_pen)

    this.mouse_up = function(){
        this.mouse_is_down = false
    }

    this.mouse_down = function(){
        this.mouse_is_down = true
    }

    this.mouse_move = function(x, y, duration){
        if (this.mouse_is_down){
            draw_line(
                this.position[0], this.position[1],          // origin
                x, y,                                        // destination
                duration,                                    // duration
                this.color[0], this.color[1], this.color[2], // color
                this.use_pen,                                // pen or eraser
                ctx                                          // context
            )
            if (this.use_pen && this.it_is_me){
                mask_shift()
            }
        }
        this.position = [x, y]
        this.reposition_label()
    }

    this.change_color = function(red, green, blue){
        this.color = [red, green, blue]
    }

    this.change_tool = function(use_pen){
        this.use_pen = use_pen
    }

    this.change_nickname = function(nickname, timestamp){
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

    this.destroy = function(){
        if (this.label){
            frame_div.removeChild(this.label)
        }
        delete users[this.user_id]
    }

    this.get_label = function(){
        return this.nickname || this.user_id
    }
}
