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
      connected do |event|
        puts "[Bot] WebSocket接続が確立されました: #{Time.now}"
      end

      # エラーが発生した時の処理
      error do |event|
        puts "[Bot] Discord接続でエラーが発生しました: #{event.message}"
        
        # 特定のエラー（Cloudflare Workers関連など）の場合は自動再接続
        if event.message.include?('WebSocket') || 
           event.message.include?('connection') ||
           event.message.include?('timeout')
          puts "[Bot] 接続エラーを検知しました。自動再接続を試行します..."
          
          Thread.new do
            sleep 10  # 10秒待機
            begin
              puts "[Bot] エラー後の再接続を開始します..."
              event.bot.run_async
              puts "[Bot] エラー後の再接続が完了しました"
            rescue StandardError => e
              puts "[Bot] エラー後の再接続に失敗しました: #{e.message}"
            end
          end
        end
      end
    end
  end
end
