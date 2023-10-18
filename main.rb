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
      logger.info("Joined VC: #{event.server.naem}(#{event.user.voice_channel.name})")
      logger.debug("Monit TC: #{event.channel.name}(#{event.channel.id})")
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
  if event.voice
    event.voice.destroy
    @text_channel.delete(event.channel.id)
    logger.info("Disconnect VC: #{event.server.name}(#{event.user.voice_channel.name})")
    logger.debug("Unmonit TC: #{event.channel.name}(#{event.channel.id})")
    event.respond('Bye!')
  else
    event.respond('ボイスチャットに参加していません。')
  end
end

# 生存確認コマンド
bot.command(:ping, { description: config.command.ping.desc, usage: config.command.ping.usage }) do |event|
  event.respond('Pong!')
end

bot.heartbeat do |event|
end

# メッセージ受信用イベント(@text_channelに入っているテキストチャンネルからのみ受信)
bot.message(start_with: not!(config.bot.prefix), in: @text_channel) do |event|
  if event.voice
    logger.info("SV: #{event.server.name}(#{event.channel.name}) USER: #{event.author.name} MSG: #{event.message.content}")
    message = event.message.content
    message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
    message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
    say(event.voice, message)
  else
    # Botがすでにチャンネルに参加していなかった場合、受信除外する
    @text_channel.delete(event.channel.id)
    logger.debug("Unmonit TC: #{event.channel.name}(#{event.channel.id})")
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
