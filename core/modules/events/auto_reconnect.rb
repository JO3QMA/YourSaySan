# frozen_string_literal: true

module YourSaySan
  module Events
    # Cloudflare Workersのアップデートによる接続断を検知して自動再接続するモジュール
    module AutoReconnect
      extend Discordrb::EventContainer

      # 接続断を検知した時の処理
      disconnected do |event|
        puts "[Bot] WebSocket接続が切断されました: #{Time.now}"
        puts "[Bot] 自動再接続を試行します..."
        
        # 少し待ってから再接続を試行
        Thread.new do
          sleep 5  # 5秒待機
          begin
            puts "[Bot] 再接続を開始します..."
            event.bot.run_async
            puts "[Bot] 再接続が完了しました"
          rescue StandardError => e
            puts "[Bot] 再接続に失敗しました: #{e.message}"
            # 再接続に失敗した場合は、さらに待ってから再試行
            sleep 30
            retry
          end
        end
      end

      # 接続が確立された時の処理
      ready do |event|
        puts "[Bot] WebSocket接続が確立されました: #{Time.now}"
      end
    end
  end
end
