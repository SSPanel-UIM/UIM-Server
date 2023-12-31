package controller

import (
	"fmt"
	"strings"

	"github.com/SSPanel-UIM/UIM-Server/api"
	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	C "github.com/sagernet/sing/common"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/shadowsocks"
	"github.com/xtls/xray-core/proxy/shadowsocks_2022"
	"github.com/xtls/xray-core/proxy/trojan"
)

func (c *Controller) buildVmessUser(userInfo *[]api.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))
	for i, user := range *userInfo {
		vmessAccount := &conf.VMessAccount{
			ID:       user.UUID,
			Security: "auto",
		}
		users[i] = &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user), // Email: InboundTag|email|uid
			Account: serial.ToTypedMessage(vmessAccount.Build()),
		}
	}
	return users
}

func (c *Controller) buildTrojanUser(userInfo *[]api.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))
	for i, user := range *userInfo {
		trojanAccount := &trojan.Account{
			Password: user.UUID,
		}
		users[i] = &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user),
			Account: serial.ToTypedMessage(trojanAccount),
		}
	}
	return users
}

func (c *Controller) buildSSUser(userInfo *[]api.UserInfo, method string) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))

	for i, user := range *userInfo {
		// shadowsocks2022 Key = "openssl rand -base64 32" and multi users needn't cipher method
		if C.Contains(shadowaead_2022.List, strings.ToLower(method)) {
			e := c.buildUserTag(&user)
			userKey, err := c.checkShadowsocksPassword(user.Passwd, method)
			if err != nil {
				newError(fmt.Errorf("[UID: %d] %s", user.UID, err)).AtError().WriteToLog()
				continue
			}
			users[i] = &protocol.User{
				Level: 0,
				Email: e,
				Account: serial.ToTypedMessage(&shadowsocks_2022.User{
					Key:   userKey,
					Email: e,
					Level: 0,
				}),
			}
		} else {
			users[i] = &protocol.User{
				Level: 0,
				Email: c.buildUserTag(&user),
				Account: serial.ToTypedMessage(&shadowsocks.Account{
					Password:   user.Passwd,
					CipherType: cipherFromString(method),
				}),
			}
		}
	}
	return users
}

func cipherFromString(c string) shadowsocks.CipherType {
	switch strings.ToLower(c) {
	case "aes-128-gcm", "aead_aes_128_gcm":
		return shadowsocks.CipherType_AES_128_GCM
	case "aes-256-gcm", "aead_aes_256_gcm":
		return shadowsocks.CipherType_AES_256_GCM
	case "chacha20-poly1305", "aead_chacha20_poly1305", "chacha20-ietf-poly1305":
		return shadowsocks.CipherType_CHACHA20_POLY1305
	case "none", "plain":
		return shadowsocks.CipherType_NONE
	default:
		return shadowsocks.CipherType_UNKNOWN
	}
}

func (c *Controller) buildUserTag(user *api.UserInfo) string {
	return fmt.Sprintf("%s|%s|%d", c.Tag, user.Email, user.UID)
}

func (c *Controller) checkShadowsocksPassword(password string, method string) (string, error) {
	return password, nil
}
