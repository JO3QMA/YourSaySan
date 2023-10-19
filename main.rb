# frozen_string_literal: true

require 'bundler/setup'
require 'discordrb'
require 'config'
require 'tempfile'
require 'logger'
require_relative './core/voicevox'

# Config
config = Config.load_and_set_settings('./config.yml', './command.yml')

# init
bot = Discordrb::Commands::CommandBot.new(
  token: config.bot.token,
  client_id: config.bot.client_id,
  prefix: config.bot.prefix,
  ignore_bots: true
)

# logging
logger = Logger.new('bot.log', 'daily')
logger.datetime_format = '%Y-%m-%d %H:%M:%S'

# Bot Init
bot.ready do
  logger.info('Main') { 'Bot is Ready.' }
  bot.game = config.bot.status
end
@text_channel = []
@voicevox = VoiceVox.new(config, logger)

# 召喚コマンド
bot.command(:summon,
            { aliases: [:s], description: config.command.summon.desc, usage: config.command.summon.usage }) do |event|
  # コマンドを実行したユーザーがVCに接続しているか確認
  if event.user.voice_channel
    # BotがVCに参加している場合は弾く
    if !event.voice
      bot.voice_connect(event.user.voice_channel)
      @text_channel << event.channel.id
      logger.info('Main') { "Joined VC: #{event.server.name}(#{event.user.voice_channel.name})" }
      logger.debug('Main') { "Monit TC: #{event.channel.name}(#{event.channel.id})" }
      event.respond('Hey!')
    else
      event.respond('すでにボイスチャットに参加しています。')
    end
  else
    event.respond('ボイスチャットに参加してから使用してください。')
  end
end

# 再生中の音声を止める
bot.command(:stop, { aliases: [:skip] }) do |event|
  if event.voice
    event.voice.stop_playing if event.voice.playing?
    nil
  else
    event.respond('ボイスチャットに参加していません。')
  end
end

# 切断コマンド
bot.command(:bye, { aliases: [:b], description: config.command.bye.desc, usage: config.command.bye.usage }) do |event|
  if event.voice_channel
    @text_channel.delete(event.channel.id)
    event.voice.destroy
    logger.info('Main') { "Disconnect VC: #{event.server.name}(#{event.user.voice_channel.name})" }
    logger.debug('Main') { "Unmonit TC: #{event.channel.name}(#{event.channel.id})" }
    event.respond('Bye!')
  else
    event.respond('ボイスチャットに参加していません。')
  end
end

# 生存確認コマンド
bot.command(:ping, { description: config.command.ping.desc, usage: config.command.ping.usage }) do |event|
  event.respond('Pong!')
end

# Debug用 evalコマンド
bot.command(:eval, help_available: false) do |event, *code|
  break unless event.user.id == 196_247_874_466_480_128

  begin
    eval code.join(' ')
  rescue StandardError
    'An error occurred ;('
  end
end

# VCに再接続するコマンド
bot.command(:reconnect) do |event|
  if event.voice
    channel = event.voice.channel
    event.voice.destroy
    bot.voice_connect(channel)
    event.respond('再接続しました。')
  else
    event.respond('ボイスチャットに参加していません。')
  end
end

# VCからユーザーが0人になった場合、自動退出
bot.voice_state_update(channel: bot.voices.values.map(&:channel)) do |event|
  unless event.voice.channel.users.map(&:current_bot).include?(false)
    @text_channel.delete(event.channel.id)
    event.voice.destroy
    logger.info('Main') { "Disconnect VC: #{event.server.name}(#{event.user.voice_channel.name})" }
    logger.debug('Main') { "Unmonit TC: #{event.channel.name}(#{event.channel.id})" }
    event.respond('Bye!')
  end
end

bot.heartbeat do |event|
end

# メッセージ受信用イベント(@text_channelに入っているテキストチャンネルからのみ受信)
bot.message(start_with: not!(config.bot.prefix), in: @text_channel) do |event|
  if event.voice
    next if event.message.content == ''

    logger.info('Main') do
      "SV: #{event.server.name}(#{event.channel.name}) USER: #{event.author.name} MSG: #{event.message.content}"
    end
    message = event.message.content
    message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
    message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
    say(event.voice, message)
  else
    # Botがすでにチャンネルに参加していなかった場合、受信除外する
    @text_channel.delete(event.channel.id)
    logger.debug('Main') { "Unmonit TC: #{event.channel.name}(#{event.channel.id})" }
    nil
  end
end

def say(voice_chat, message)
  Tempfile.create do |tempfile|
    sound = @voicevox.speak(message)
    tempfile.write(sound)
    voice_chat.play_file(tempfile.path)
  end
end

bot.run
