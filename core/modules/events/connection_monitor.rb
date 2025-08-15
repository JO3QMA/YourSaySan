# frozen_string_literal: true

module YourSaySan
  module Events
    # 接続状態を定期的に監視して自動再接続するモジュール
    module ConnectionMonitor
      extend Discordrb::EventContainer

      def self.setup(bot, text_channels, voicevox, config)
        @bot = bot
        @config = config
        @last_heartbeat = Time.now
        @monitoring = false
        
        # 自動再接続が無効の場合は何もしない
        return unless config.bot.auto_reconnect&.enabled
        
        # ハートビートイベントを監視
        heartbeat do |event|
          @last_heartbeat = Time.now
        end

        # 接続監視スレッドを開始
        start_monitoring unless @monitoring
      end

      private

      def self.start_monitoring
        @monitoring = true
        
        Thread.new do
          loop do
            sleep @config.bot.auto_reconnect.check_interval
            
            begin
              # 最後のハートビートから設定された時間以上経過している場合は再接続
              if Time.now - @last_heartbeat > @config.bot.auto_reconnect.heartbeat_timeout
                puts "[Bot] ハートビートが#{@config.bot.auto_reconnect.heartbeat_timeout}秒以上途絶えています。再接続を試行します..."
                reconnect_bot
              end
              
              # Botの接続状態をチェック
              unless @bot.connected?
                puts "[Bot] Botが切断状態です。再接続を試行します..."
                reconnect_bot
              end
              
            rescue StandardError => e
              puts "[Bot] 接続監視中にエラーが発生しました: #{e.message}"
            end
          end
        end
      end

      def self.reconnect_bot
        begin
          puts "[Bot] 自動再接続を開始します..."
          
          # 既存の接続をクリーンアップ
          @bot.stop if @bot.connected?
          
          # 少し待ってから再接続
          sleep 5
          
          # 非同期で再接続
          @bot.run_async
          
          puts "[Bot] 自動再接続が完了しました"
          
        rescue StandardError => e
          puts "[Bot] 自動再接続に失敗しました: #{e.message}"
          
          # 失敗した場合は設定された時間後に再試行
          Thread.new do
            sleep @config.bot.auto_reconnect.retry_delay
            reconnect_bot
          end
        end
      end
    end
  end
end
