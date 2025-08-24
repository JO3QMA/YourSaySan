# frozen_string_literal: true

module YourSaySan
  module Events
    # Cloudflare Workersのアップデートによる接続断を検知して自動再接続するモジュール
    module AutoReconnect
      extend Discordrb::EventContainer

      # 接続断を検知した時の処理
      disconnected do |_event|
        puts "[Bot] WebSocket接続が切断されました: #{Time.now}"
        puts '[Bot] Discordrbの自動再接続機能が動作します...'
        # Discordrbが自動的に再接続を行うため、手動での再接続処理は不要
      end

      # 接続が確立された時の処理
      ready do |_event|
        puts "[Bot] WebSocket接続が確立されました: #{Time.now}"
      end
    end
  end
end
