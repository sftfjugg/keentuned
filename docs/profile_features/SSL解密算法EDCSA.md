# 优化原理
ECDSA是用于数字签名，是ECC与DSA的结合，整个签名过程与DSA类似，所不一样的是签名中采取的算法为ECC，最后签名出来的值也是分为r,s。而ECC（全称Elliptic Curves Cryptography）是一种椭圆曲线密码编码学。
ECDH每次用一个固定的DH key，导致不能向前保密（forward secrecy），所以一般都是用ECDHE（ephemeral）或其他版本的ECDH算法。ECDH则是基于ECC的DH（ Diffie-Hellman）密钥交换算法。
ECC与RSA 相比，有以下的优点：

- a. 相同密钥长度下，安全性能更高，如160位ECC已经与1024位RSA、DSA有相同的安全强度。
- b. 计算量小，处理速度快，在私钥的处理速度上（解密和签名），ECC远 比RSA、DSA快得多。
- c. 存储空间占用小 ECC的密钥尺寸和系统参数与RSA、DSA相比要小得多， 所以占用的存储空间小得多。
- d. 带宽要求低使得ECC具有广泛得应用前景。
# 使用方法

1. 生成秘钥和证书
```
mkdir /etc/pki/nginx/  /etc/pki/nginx/private -p
openssl genrsa -des3 -out server.key 2048  #会有两次要求输入密码,输入同一个即可
openssl rsa -in server.key -out server.key
openssl req -new -key server.key -out server.csr
openssl req -new -x509 -key server.key -out server.crt -days 3650
openssl req -new -x509 -key server.key -out ca.crt -days 3650
openssl x509 -req -days 3650 -in server.csr -CA ca.crt -CAkey server.key -CAcreateserial -out server.crt

cp server.crt /etc/pki/nginx/
cp server.key /etc/pki/nginx/private
```

2. 启用nginx.conf中的https相关配置
```
sed -i "57,81s/#\(.*\)/\1/" /etc/nginx/nginx.conf
```

3. 配置EDCSA证书
```
# ……
    server {
        listen       443 ssl http2;
        listen       [::]:443 ssl http2;
        server_name  _;
        root         /usr/share/nginx/html;

        ssl_certificate "/etc/pki/nginx/server.crt";
        ssl_certificate_key "/etc/pki/nginx/private/server.key";
        ssl_session_cache shared:SSL:1m;
        ssl_session_timeout  10m;
        ssl_ciphers PROFILE=SYSTEM;
        ssl_prefer_server_ciphers on;

        # Load configuration files for the default server block.
        include /etc/nginx/default.d/*.conf;

        error_page 404 /404.html;
            location = /40x.html {
        }

        error_page 500 502 503 504 /50x.html;
            location = /50x.html {
        }
    }
# ……
```
