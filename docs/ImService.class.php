<?php

class ImService
{
    /**
     * @var ImService
     */
    private static $instance;

    /**
     * socket对象
     *
     * @var
     */
    private $socket;
    /**
     * 长连接服务器地址
     *
     * @var string
     */
    private $host;
    /**
     * 长连接服务器端口
     *
     * @var string
     */
    private $port;
    /**
     * 系统消息私钥
     *
     * @var string
     */
    private $systemKey;
    /**
     * 是否调试模式
     */
    private $debug = false;

    const IM_FROM_TYPE_SYSTEM = 1;   // 消息类型
    const IM_CHAT_BORADCAST   = 401; // 世界消息
    const IM_CHAT_GROUP       = 402; // 频道消息
    const IM_CHAT_PRIVATE     = 403; // 私聊消息

    const IM_STAT            = 101; // 统计服务器状态
    const IM_CHECK_ONLINE    = 102; // 判断用户是否在线
    const IM_KICK_USER       = 103; // 踢某用户下线
    const IM_KICK_ALL        = 104; // 踢所有用户下线
    const IM_GROUP_USER_LIST = 105; // 获取频道用户列表

    /**
     * 消息协议中长度部分的长度
     */
    const IM_LENGTH_SIZE = 2;
    const IM_HEADER_SIZE = 4;
    const IM_TYPE_SIZE = 2;
    const IM_FROM_TYPE_SIZE = 2;
    const IM_BODY_MAX_LENGTH = 2048;

    /**
     * Get the instance of ImService.
     *
     * @param string $host
     * @param string $port
     * @param string $systemKey
     * @return ImService
     */
    public static function get($host = "127.0.0.1", $port = "8899", $systemKey = "XXOOOOXX")
    {
        if (!self::$instance) {
            self::$instance = new ImService($host, $port, $systemKey);
        }
        return self::$instance;
    }

    /**
     * Construction.
     *
     * @param string $host
     * @param string $port
     * @param string $systemKey
     */
    private function __construct($host, $port, $systemKey)
    {
        $this->host = $host;
        $this->port = $port;
        $this->systemKey = $systemKey;
    }

    /**
     * 调试模式开关
     *
     * @param boolean $debug default true
     */
    public function setDebug($debug = true)
    {
        $this->debug = $debug;
    }

    /**
     * 生成登录密钥
     *
     * @param int $userID
     * @param int $time
     * @return string
     */
    public function getLoginToken($userID, $time)
    {
        return substr(md5(sprintf('%s#%d#%d',$this->loginKey, $userID, $time)), 0, 4);
    }

    /**
     * 连接socket
     */
    public function connect()
    {
        $this->dump("Start to create socket..." );

        $this->socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
        if (false === $this->socket) {
            $this->dump("socket_create() failed: reason: ".socket_strerror(socket_last_error()));
        } else {
            $this->dump("socket_create OK.");
        }

        $this->dump("Attempting to connect to " . $this->host . " on port " .$this->port . "...");

        $result = socket_connect($this->socket, $this->host, $this->port);
        if (false === $result) {
            $this->dump("socket_connect() failed.");
            $this->dump("Reason: ($result) " . socket_strerror(socket_last_error()));
        } else {
            $this->dump("socket_connect OK.");
        }
    }

    /**
     * 关闭socket连接
     */
    public function close()
    {
        socket_close($this->socket);
    }

    /**
     * 发送消息
     *
     * @param int $imType
     * @param array $body default array
     * @return boolean
     */
    private function send($imType, $body = array())
    {
        // 协议数据
        $socketData = "";

        // 写入系统秘钥
        $body['systemKey'] = $this->systemKey;
        $jsonData = json_encode($body);

        // 包体长度
        $bodyLength = strlen($jsonData);
        // 协议长度，包体长度+协议长度
        $length = $bodyLength + self::IM_HEADER_SIZE;

        // 生成包头
        $socketData .= pack("n", $length);
        $socketData .= pack("n", $imType);
        $socketData .= pack("n", self::IM_FROM_TYPE_SYSTEM);

        // 生成包体
        $socketData .= pack("a" . $bodyLength, $jsonData);

        // 发送消息
        if (false === @socket_write($this->socket, $socketData, $length + self::IM_LENGTH_SIZE)) {
            $this->dump("socket_write() failed: reason: ".socket_strerror(socket_last_error()));
            return false;
        } else {
            $this->dump("socket_write() Ok.");
            return true;
        }
    }

    /**
     * 接受socket消息
     */
    private function read()
    {
        $socketData = array();

        // 读取协议长度
        $length = current(unpack("n", socket_read($this->socket, self::IM_LENGTH_SIZE)));
        $socketData['length'] = $length;
        // 读取消息类型
        $socketData['imType'] = current(unpack("n", socket_read($this->socket, self::IM_TYPE_SIZE)));
        // 读取来源类型
        $socketData['fromType'] = current(unpack("n", socket_read($this->socket, self::IM_FROM_TYPE_SIZE)));
        // 读取协议内容
        $socketData['body'] = json_decode(socket_read($this->socket, $length-self::IM_HEADER_SIZE));

        return $socketData;
    }

    /**
     * 发送系统消息
     *
     * @param array $body
     * @return boolean
     */
    public function sendSystem($body)
    {
        $this->dump("sendSystem:" . var_export($body, true));
        $sendResult = $this->send(self::IM_CHAT_BORADCAST, $body);
        if ($sendResult) {
            $response = $this->read();
            $this->dump("sendSystem Response:" . var_export($response, true));
        }

        return $sendResult;
    }

    /**
     * 发送频道消息
     *
     * @param string $groupID
     * @param array $body default array
     * @return boolean
     */
    public function sendGroup($groupID, $body = array())
    {
        $this->dump("sendGroup: groupID:" . $groupID . ", " . var_export($body, true));
        $body['groupID'] = strval($groupID);
        $sendResult = $this->send(self::IM_CHAT_GROUP, $body);
        if ($sendResult) {
            $response = $this->read();
            $this->dump("sendGroup Response:" . var_export($response, true));
        }

        return $sendResult;
    }

    /**
     * 发送私聊消息
     *
     * @param int $userID
     * @param array $body
     * @return boolean
     */
    public function sendPrivate($userID, $body = array())
    {
        $this->dump("sendGroup: userID:" . $userID . ", " . var_export($body, true));
        $body['receiverId'] = strval($userID);
        $sendResult = $this->send(self::IM_CHAT_PRIVATE, $body);
        if ($sendResult) {
            $response = $this->read();
            $this->dump("sendPrivate Response:" . var_export($response, true));
        }

        return $sendResult;
    }

    /**
     * 判断用户是否在线
     *
     * @param int $userID
     * @return int
     */
    public function checkOnline($userID)
    {
        $online = 0;
        $body = array('userID' => strval($userID));
        if ($this->send(self::IM_CHECK_ONLINE, $body)) {
            $response = $this->read();
            $this->dump("checkOnline Response:" . var_export($response, true));
            $online = $response['body']->onLine;
        }

        return $online;
    }

    /**
     * 获取统计信息
     *
     * @return array
     */
    public function getStat()
    {
        $statInfo = array();
        if ($this->send(self::IM_STAT, array())) {
            $response = $this->read();
            $this->dump("getStat Response:" . var_export($response, true));
            $statInfo = $response['body'];
        }

        return $statInfo;
    }

    /**
     * 踢某用户下线
     *
     * @param int $userID
     * @param string $reason
     * @return boolean
     */
    public function kickUser($userID, $reason)
    {
        $sendResult = $this->send(self::IM_KICK_USER, array('userID' => strval($userID), 'msg' => $reason));
        if ($sendResult) {
            $response = $this->read();
            $this->dump("kickUser Response:" . var_export($response, true));
        }

        return $sendResult;
    }

    /**
     * 踢所有用户下线
     *
     * @param string $reason
     * @return boolean
     */
    public function kickAll($reason)
    {
        $sendResult = $this->send(self::IM_KICK_ALL, array('msg' => $reason));
        if ($sendResult) {
            $response = $this->read();
            $this->dump("kickAll Response:" . var_export($response, true));
        }

        return $sendResult;
    }

    /**
     * 获取频道用户列表
     *
     * @param string $groupID
     * @return array
     */
    public function getGroupUserList($groupID)
    {
        $groupUserList = array();
        if ($this->send(self::IM_GROUP_USER_LIST, array('groupID' => strval($groupID)))) {
            $response = $this->read();
            $this->dump("getGroupUserList Response:" . var_export($response, true));
            $groupUserList = $response['body']->userList;
        }

        return $groupUserList;

    }

    /**
     * socket输出内容
     *
     * @param string 输出内容
     */
    private function dump($msg)
    {
        if ($this->debug) {
            echo "[" . date('Y-m-d H:i:s') . "] " . $msg . "<br />" . PHP_EOL;
        }
    }
}