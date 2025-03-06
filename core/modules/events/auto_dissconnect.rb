# frozen_string_literal: true

module YourSaySan
  module Events
    # VCにBotしか存在しない場合はVCから切断する
    module AutoDissconnect
      extend Discordrb::EventContainer

      voice_state_update do |event|
        # ユーザーがVCから切断した場合、event.channelはnilを返す
        # BotがVCに参加してない場合、bot.voice(event.server)はnilを返す
        if event.channel.nil?
          # 切断したVCはevent.old_channel
          puts "[Bot] 切断検知: #{event.old_channel.name}"
          event.old_channel.users.each do |user|
            puts "残りユーザー: #{user.name} #{user.bot_account}"
          end

          vc_user_list = event.old_channel.users
          if vc_user_list.map(&:bot_account).count(false).zero? && vc_user_list.any? do |user|
            user.id == event.bot.bot_user.id
          end
            YourSaySan::BOT.voice_destroy(event.server)
          end
        end
      end
    end
  end
end
