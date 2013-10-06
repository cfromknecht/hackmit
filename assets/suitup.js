var su = ''; //http://ec2-54-200-40-68.us-west-2.compute.amazonaws.com:8080';

var _user_details = '/user/details';

var chat_check_id, firepad, chatroomid;

var _user_details = 'http://web.mit.edu/ambhave/www/suitup/user_details.json';

function init_user_details() {
    console.log('init_user_details()');

    $.getJSON(_user_details, function (data) {
        $('.user_username').text(data['username']);
    }).error(function (jqXhr, textStatus, error) {
        alert("ERROR: " + textStatus + ", " + error);
    });
}

$(document).ready(function () {
  //  login();
    //  init_user_details();
    
});

function chat_join(s) {
    $.ajax({
        type: "GET",
        url: su + '/chatroom/join',
        success: function (data) {
            chatroomid = JSON.parse(data).crid;
            $('#firepad').html('');
            var firepadRef = new Firebase('//hackmitsuitup.firebaseIO.com/firepads/' + chatroomid);
            var codeMirror = CodeMirror(document.getElementById('firepad'), {
                lineNumbers: true,
                mode: 'python'
            });
            firepad = Firepad.fromCodeMirror(firepadRef, codeMirror);
            webrtc.on('readyToCall', function () {
                $('video').css('height', '200px');
            });

            setTimeout(join_room, 2000);
            chat_check_id = setTimeout(chat_check, 1000);
            //question_new();

            $('.questions_title').text(JSON.parse(data).question.title);
            $('.questions_body').html(JSON.parse(data).question.body);
        }
    });
}

function join_room() {
    // webrtc.createRoom(chatroomid);
    webrtc.joinRoom(String(chatroomid));
}

function chat_leave() {
    clearTimeout(chat_check_id);
    $.ajax({
        type: "GET",
        url: su + '/chatroom/leave',
        success: function (data) {
            $('#firepad').html('');
        }
    });
    chatroomid = null;
    webrtc.leaveRoom(webrtc.roomName);
}
$(window).bind('beforeunload', function () {
    $.ajax({
        type: "GET",
        url: su + '/chatroom/leave',
        async: false        
    });
    
});

function chat_send(chat, convo) {
    $.ajax({
        type: "GET",
        url: su + '/message/send',
        data: { 's': $('#' + chat).val() }        
    });
    $('#' + convo).html($('#' + convo).html() + '<br /><strong style="color:blue">Me:</strong> ' + $('#' + chat).val());
    $('#' + chat).val('');    
}

function chat_check() {
    $.ajax({
        type: "GET",
        url: su + '/message/check',
        success: function (data) {
            sts = JSON.parse(data).status;
            if ( sts == "failure") {
                console.log("Rejoining to new person");
                $('#firepad').html('');
                $('#convo').val('');
                $('#chat').val('');
                chat_join();
            } else {
                if(JSON.parse(data).s != "") {
                    //$('#convo').val($('#convo').val() + '\nOther: ' + JSON.parse(data).s);
                    $('#convo').html($('#convo').html() + '<br /><strong style="color:red">Other:</strong> ' + JSON.parse(data).s);
                    setTimeout(chat_check, 1000);
                } else {
                    setTimeout(chat_check, 2000);
                }
            }
        }
    });
}







function question_new() {
    $.getJSON(su + '/question/new', {'cvid': chatroomid }, function (data) {
        $('.questions_title').text(data['title']);
        $('.questions_body').html(data['body']);
    }).error(function (jqXhr, textStatus, error) {
        alert("ERROR: " + textStatus + ", " + error);
    });
}


function question_submit() {
    console.log("submit");
    $.ajax({
        type: "POST",
        data: {'submission': firepad.getText(), 'cvid': chatroomid },
        url: su + '/question/submit',
        success: function (data) {
            // console.log(data);
            // if (JSON.parse(data).status == "success") {
            //     dt = JSON.parse(data).data;
            //     $('#convo').html($('#convo').html() + '<br /><strong style="color:red">System:</strong> ' + dt);
            //     // $('#convo').html($('#convo').html() + '<br /><strong style="color:black">System:</strong> ' + dt;
            // }
            $('#convo').html($('#convo').html() + '<br /><strong style="color:black">System:</strong> ' + data);
        }
    });
}



// Load the SDK asynchronously
(function (d) {
    var js, id = 'facebook-jssdk', ref = d.getElementsByTagName('script')[0];
    if (d.getElementById(id)) { return; }
    js = d.createElement('script'); js.id = id; js.async = true;
    js.src = "//connect.facebook.net/en_US/all.js";
    ref.parentNode.insertBefore(js, ref);
}(document));


var access_token = null;
var user_city = null;
var fb_user = null;

function login() {
    _login_wait();

    FB.login(function (response) {
        //_check_login();      
    });
}
function logout() {
    _login_wait();

    FB.logout(function (response) {
        //_login_bad();
    });
}

// Additional JS functions here
window.fbAsyncInit = function () {
    FB.init({
        appId: '615988445111525', // App ID
        channelUrl: '//localhost:8000/channel.html', // Channel File
        status: true, // check login status
        cookie: true, // enable cookies to allow the server to access the session
        xfbml: true  // parse XFBML
    });

    // Additional init code here
    //_check_login();
    _login_wait();
    FB.getLoginStatus(function (response) {
        if (response.status !== 'connected') {
            _login_bad();
        }
    });

    FB.Event.subscribe('auth.statusChange', function (response) {
        _check_login();
        console.log(response);
    });
};

/* Helper Functions */
function _hide_all_login() {
    $('.pLoginBad').hide();
    $('.pLoginDone').hide();
    $('.pLoginWait').hide();
}
function _check_login() {
    _login_wait();
    FB.getLoginStatus(function (response) {
        if (response.status === 'connected') {
            _login_done(response);
        }
        else {
            _login_bad();
        }
    });
}
function _login_wait() {
    _hide_all_login();
    $('.pLoginWait').show();
}
function _login_bad() {
    _hide_all_login();
    $('.pLoginBad').show();
}
function _login_done(response) {

    if (response != null)
        access_token = response.authResponse.accessToken;

    FB.api('/me?fields=picture,name', function (response) {
        fb_user = response;
        console.log(response);
        $('.user_username').text(response.name);
        $('.user_photourl').attr('src', response.picture.data.url);

        _hide_all_login();
        $('.pLoginDone').show();
    });


    // TODO: UNCOMMENT THIS!!!!!!!!!!!!!!!!!!!!!!!!!!!
    $.ajax({ url: su + '/login?access_token=' + access_token }).success(function (data) { chat_join(); });



    //$.ajax({url:"demo_test.txt",success:function(result){    $("#div1").html(result);  }});

    /*
    
    FB.api('/me/home', function(response) {
       console.log(response);
       $('#news_feed').empty();
       for (var i = 0; i < response['data'].length; i++) {
           var item = response['data'][i];
           var txt = '<div class="media"> \
                   <a class="pull-left" href="#"> \
                       <img class="media-object" src="' + item['icon'] + '"> \
                   </a> \
                   <div class="media-body">  \
                       <h4 class="media-heading"><a href="#">' + item['from']['name'] + '</a></h4>            \
                   </div> \
               </div>';
           $('#news_feed').append(txt);
       }
    });*/
}
function _fatal_error(message) {
    $('body').html('Fatal Error! Please refresh the page<br /><br />Message: ' + message);
}

