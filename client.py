import os, sys

try:
    if os.name == 'nt':
        from twisted.internet import iocpreactor
        iocpreactor.install()
    else:
        from twisted.internet import epollreactor
        epollreactor.install()
except:
    print 'Failed to install OS specific reactor'

from twisted.internet import protocol, reactor, defer, utils
from twisted.protocols import basic
from twisted.web import client
from twisted.enterprise import adbapi

from settings import db #DB_DRIVER, DB_HOST, DB_USER, DB_PASS, DB_NAME

dbpool = adbapi.ConnectionPool(db.DRIVER, db.HOST, db.USER, db.PASS, db.NAME)

def catchError(err):
    print 'Error: ', str(err)
    reactor.stop()

def getSongsFromHtml(html):
    startToken = html.find("1) ")
    endToken = html.find('<br/></div>')
    items = html[startToken:endToken]
    items = items.split('<br/>')
    songs = []
    for item in items:
        bracketPos = item.find(') ')
        songs.append(item[bracketPos+2:])
    print 'Got playlist: ', songs
    return songs

def getLastSong():
    return dbpool.runQuery("SELECT * FROM playlist ORDER BY id DESC LIMIT 1")

def lastSongSuccess(song):
    if len(song) == 0:
        print 'No last song found'
        return ''
    else:
        print 'Got last song: ', song
        return song[0][1]

def lastSongFailure(err):
    print 'Failed to get last song: ', err

def doIt(results):
    lastSongResult, lastSong = results[0]
    if lastSongResult == False:
        lastSong = ''

    playlistResult, songs = results[1]
    if playlistResult == True:
        i = len(songs)
        try:
            i = songs.index(lastSong)
        except ValueError as e:
            pass

        questionList = []
        insertList = []
        for index in range(i-1, -1, -1):
            song = songs[index]
            insertList.append(song)
            questionList.append("(%s)")

        if len(insertList) > 0:
            d2 = dbpool.runQuery("INSERT INTO playlist (name) VALUES " + ','.join(questionList) + "", tuple(insertList))
            def addSongsSuccess(result):
                print 'Songs added: ', insertList
                print result
                reactor.stop()
            def addSongsFailure(result):
                print 'Failed to add songs:', insertList
                print result
                reactor.stop()
                #sys.exit(1)
            d2.addCallback(addSongsSuccess)
            d2.addErrback(addSongsFailure)
        else:
            print 'No songs to add'
            reactor.stop()
    else:
        print 'Couldn\'t retrieve playlist: ', songs
        reactor.stop()
        #sys.exit(1)

if __name__ == "__main__":
    dLastSong = getLastSong()
    dLastSong.addCallbacks(lastSongSuccess, lastSongFailure)

    dPlaylist = client.getPage('http://www.tuksfm.co.za/PlayList.aspx')
    dPlaylist.addCallbacks(getSongsFromHtml, catchError)

    dl = defer.DeferredList([dLastSong, dPlaylist])
    dl.addCallbacks(doIt, catchError)

    reactor.run()
