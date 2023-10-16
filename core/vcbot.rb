# frozen_string_literal: true

require 'discordrb'
require 'tempfile'
require 'opus-ruby'
require_relative './voicevox'

# VCに繋いだ時の受け手
class VCBot
  attr_accessor :name

  def initialize(config, event)
    bot_init(config)
    @text_channel = event.channel
    @voicevox = VoiceVox.new
    @bot.ready do
      @bot.voice_connect(event.user.voice_channel)
    end
  end

  def bot_init(config)
    @bot = Discordrb::Bot.new(
      token: config.bot.token,
      client_id: config.bot.client_id,
      ignore_bots: true
    )
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

  def kill
    @bot.stop
  end

  def main
    @bot.message(in: @text_channel) do |event|
      next if event.author.bot_account?
      next if event.message.content.start_with?('!')

      if event.message.content =~ URI::DEFAULT_PARSER.make_regexp
        speak(event, 'URL省略')
        next
      end

      if event.message.content.size >= 50
        message = "#{event.message.content[0, 49]} + 以下略"
        speak(event, message)
      end

      puts "SV: #{event.server.name}(#{event.channel.name}) USER: #{event.author.name} MSG: #{event.message.content}"
      speak(event, event.message.content)
    end
    @bot.run
  end
end
