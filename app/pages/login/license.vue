<template>
  <view class="page-login-license">
    <DsNavBar title="用户服务协议"/>
    <view class="license-title">用户服务协议</view>
    
    <view class="license-section">
      <text class="section-title">一、协议概述</text>
      <text class="section-content">
        欢迎使用{{ systemName }}物联网平台（以下简称"本平台"）。本协议是您与{{ company || systemName }}运营方（以下简称"我们"）之间关于使用本平台服务的法律协议。
        您在使用本平台前，应仔细阅读并理解本协议的全部内容。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">二、服务内容</text>
      <text class="section-content">
        本平台提供物联网设备管理、远程控制、数据采集与分析、智能场景联动等服务。
        您可以通过本平台连接和管理各类物联网设备，包括但不限于传感器、智能开关、温湿度计、摄像头等。
        我们有权根据业务发展需要调整服务内容，并将通过系统公告及时通知您。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">三、账号注册与使用</text>
      <text class="section-content">
        您注册账号时应提供真实、准确的信息。您应对账号安全负责，妥善保管登录凭证，不得将账号转借他人使用。
        如发现账号异常，应立即修改密码并联系我们进行处理。我们有权对异常账号采取冻结、封禁等措施。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">四、设备管理</text>
      <text class="section-content">
        您可以通过本平台添加、删除和管理物联网设备。添加设备时，您需要确保拥有该设备的合法所有权或使用权。
        您应对设备的使用行为负责，不得利用设备从事违法违规活动。如设备出现故障，您应及时排查或联系设备厂商。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">五、数据与隐私</text>
      <text class="section-content">
        本平台会采集设备上传的数据（包括但不限于设备状态、传感器数据、控制指令等）用于提供服务。
        我们重视您的数据安全和隐私保护，将按照隐私政策严格保护您的个人信息和设备数据。
        未经您授权，我们不会向第三方披露您的个人信息和设备数据。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">六、服务费用</text>
      <text class="section-content">
        本平台基础服务免费提供。部分增值服务（如高级数据分析、API调用、设备接入数量超限等）可能收取费用，
        收费标准将在服务页面明确显示。您使用付费服务即表示同意支付相应费用。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">七、用户义务</text>
      <text class="section-content">
        您在使用服务过程中，应遵守中华人民共和国法律法规，不得利用本平台从事违法违规活动。
        不得干扰平台正常运行，不得恶意攻击系统，不得未经授权访问他人设备。
        应妥善保管设备，确保设备网络连接稳定，避免因设备离线导致服务中断。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">八、安全责任</text>
      <text class="section-content">
        您应采取必要的安全措施保护设备和数据安全，包括定期更新设备固件、设置强密码、开启双重认证等。
        因您未采取足够安全措施导致的设备被盗、数据泄露等损失，由您自行承担。
        如发现安全漏洞或可疑活动，请及时通知我们。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">九、协议变更</text>
      <text class="section-content">
        我们有权根据业务发展需要修改本协议。修改后的协议将通过系统公告发布，
        您继续使用服务即表示同意接受修改后的协议。如您不同意修改，应停止使用本平台。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">十、服务终止</text>
      <text class="section-content">
        您可以随时注销账号终止服务。我们有权在以下情况下终止服务：违反本协议约定、
        从事违法违规活动、长期不活跃账号等。服务终止后，我们将停止保存您的设备数据，
        请提前做好数据备份。
      </text>
    </view>

    <view class="license-section">
      <text class="section-title">十一、争议解决</text>
      <text class="section-content">
        因本协议产生的争议，双方应友好协商解决。协商不成的，可向我们所在地人民法院提起诉讼。
        本协议的解释权归我们所有。
      </text>
    </view>

    <view class="license-footer">
      <text class="footer-text">{{ systemName }}物联网平台</text>
      <text class="footer-text">最后更新日期：2026年7月</text>
    </view>
  </view>
</template>

<script setup>
import { getSystemName, getOEM } from '@/utils/oem'
import DsNavBar from '@/components/DsNavBar.vue';

const systemName = getSystemName()
const { company } = getOEM()
</script>

<style lang="scss">
page {
  background: #F0F7FF;
}

.license-container {
  min-height: 100vh;
  background: #F0F7FF;
  padding-bottom: 100rpx;
}

.license-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #1E3A5F;
  text-align: center;
  margin-bottom: 40rpx;
}

.license-section {
  margin-bottom: 30rpx;
}

.section-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #5BA3E8;
  display: block;
  margin-bottom: 16rpx;
}

.section-content {
  font-size: 26rpx;
  color: #6B7280;
  line-height: 1.8;
}

.license-footer {
  text-align: center;
  padding-top: 40rpx;
  margin-top: 20rpx;
  border-top: 1rpx solid #D1E7FD;
}

.footer-text {
  font-size: 24rpx;
  color: #9CA3AF;
  display: block;
  margin-bottom: 8rpx;
}
</style>
