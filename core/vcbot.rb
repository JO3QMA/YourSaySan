# frozen_string_literal: true

require 'discordrb'
require 'tempfile'
require_relative './voicevox'

# VCに繋いだ時の受け手
class VCBot
  def initialize(bot, event)
    @bot = bot
    @text_channel = event.channel
    @voice_channel = event.user.voice_channel
    @bot.voice_connect(@voice_channel)
    @voicevox = VoiceVox.new
  end

  def speak(event, message)
    query = @voicevox.voice_query(message)
    sound = @voicevox.speak(query)
    temp = tempvoice(sound)
    event.voice.play_file(temp.path)
    temp.unlink
  end

  def tempvoice(sound)
    tempfile = Tempfile.new('voice', encording: 'UTF-8')
    tempfile.write(sound)
    tempfile.close

    tempfile
  end

  def main
    @bot.message(in: @text_channel) do |event|
      next if event.author.bot_account?
      next if event.message.content.start_with?('!')

      if event.message.content =~ URI::DEFAULT_PARSER.make_regexp
        speak(event, 'URL省略')
        next
      end

      puts "サーバー: #{event.server.name} チャンネル: #{event.channel.name} ユーザー: #{event.author.name} メッセージ: #{event.message.content}"
      speak(event, event.message.content)
    end
  end
end
